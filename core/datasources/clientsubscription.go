package datasources

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
)

// ClientSubscriptionBufferSize is the buffer size of the subscription channel.
// It is used to prevent blocking the client dispatcher when the client is slow to consume values.
var ClientSubscriptionBufferSize = 8

// ClientSubscription is a subscription to a stream of values from the client dispatcher.
// It has two channels: one for values, and one for errors.
type ClientSubscription[T any] struct {
	// The channel which the subscription sends values.
	channel chan T

	// The in channel receives values from client dispatcher.
	in chan T

	// The error channel receives the error from the client dispatcher.
	err       chan error
	quiteOnce sync.Once

	// Closing of the subscription is requested by sending on 'quit'. This is handled by
	// the forwarding loop, which closes 'forwardDone' when it has stopped sending to
	// sub.channel. Finally, 'unsubDone' is closed after unsubscribing on the server side.
	quit     chan struct{}
	quitDone chan struct{}
}

func newClientSubscription[T any](channel chan T) *ClientSubscription[T] {
	return &ClientSubscription[T]{
		channel:  channel,
		in:       make(chan T, ClientSubscriptionBufferSize),
		err:      make(chan error, ClientSubscriptionBufferSize),
		quit:     make(chan struct{}),
		quitDone: make(chan struct{}),
	}
}

func (c *ClientSubscription[T]) Unsubscribe() {
	_ = c.UnsubscribeWithContext(context.Background())
}

func (c *ClientSubscription[T]) UnsubscribeWithContext(ctx context.Context) (err error) {
	c.quiteOnce.Do(func() {
		select {
		case c.quit <- struct{}{}:
			<-c.quitDone
		case <-ctx.Done():
			err = ctx.Err()
		}
	})
	return errors.WithStack(err)
}

// Err returns the error channel of the subscription.
func (c *ClientSubscription[T]) Err() <-chan error {
	return c.err
}

// send sends a value to the subscription channel. If the subscription is closed, it returns an error.
func (c *ClientSubscription[T]) send(ctx context.Context, value T) error {
	select {
	case c.in <- value:
	case <-c.quitDone:
		return errors.Wrap(errs.InternalError, "subscription is closed")
	case <-ctx.Done():
		return errors.WithStack(ctx.Err())
	}
	return nil
}

// sendError sends an error to the subscription error channel. If the subscription is closed, it returns an error.
func (c *ClientSubscription[T]) sendError(ctx context.Context, err error) error {
	select {
	case c.err <- err:
	case <-c.quitDone:
		return errors.Wrap(errs.InternalError, "subscription is closed")
	case <-ctx.Done():
		return errors.WithStack(ctx.Err())
	}
	return nil
}

// run starts the forwarding loop for the subscription.
func (c *ClientSubscription[T]) run() {
	defer close(c.err)
	defer close(c.quitDone)

	for {
		select {
		case <-c.quit:
			return
		case value := <-c.in:
			select {
			case c.channel <- value:
			case <-c.quit:
				return
			}
		}
	}
}
