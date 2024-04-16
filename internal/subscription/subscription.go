package subscription

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
)

// SubscriptionBufferSize is the buffer size of the subscription channel.
// It is used to prevent blocking the client dispatcher when the client is slow to consume values.
var SubscriptionBufferSize = 8

// Subscription is a subscription to a stream of values from the client dispatcher.
// It has two channels: one for values, and one for errors.
type Subscription[T any] struct {
	// The channel which the subscription sends values.
	channel chan<- T

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

func NewSubscription[T any](channel chan<- T) *Subscription[T] {
	subscription := &Subscription[T]{
		channel:  channel,
		in:       make(chan T, SubscriptionBufferSize),
		err:      make(chan error, SubscriptionBufferSize),
		quit:     make(chan struct{}),
		quitDone: make(chan struct{}),
	}
	go func() {
		subscription.run()
	}()
	return subscription
}

func (s *Subscription[T]) Unsubscribe() {
	_ = s.UnsubscribeWithContext(context.Background())
}

func (s *Subscription[T]) UnsubscribeWithContext(ctx context.Context) (err error) {
	s.quiteOnce.Do(func() {
		select {
		case s.quit <- struct{}{}:
			<-s.quitDone
		case <-ctx.Done():
			err = ctx.Err()
		}
	})
	return errors.WithStack(err)
}

// Client returns a client subscription for this subscription.
func (s *Subscription[T]) Client() *ClientSubscription[T] {
	return &ClientSubscription[T]{
		subscription: s,
	}
}

// Err returns the error channel of the subscription.
func (s *Subscription[T]) Err() <-chan error {
	return s.err
}

// Done returns the done channel of the subscription
func (s *Subscription[T]) Done() <-chan struct{} {
	return s.quitDone
}

// IsClosed returns status of the subscription
func (s *Subscription[T]) IsClosed() bool {
	select {
	case <-s.quitDone:
		return true
	default:
		return false
	}
}

// Send sends a value to the subscription channel. If the subscription is closed, it returns an error.
func (s *Subscription[T]) Send(ctx context.Context, value T) error {
	select {
	case s.in <- value:
	case <-s.quitDone:
		return errors.Wrap(errs.InternalError, "subscription is closed")
	case <-ctx.Done():
		return errors.WithStack(ctx.Err())
	}
	return nil
}

// SendError sends an error to the subscription error channel. If the subscription is closed, it returns an error.
func (s *Subscription[T]) SendError(ctx context.Context, err error) error {
	select {
	case s.err <- err:
	case <-s.quitDone:
		return errors.Wrap(errs.InternalError, "subscription is closed")
	case <-ctx.Done():
		return errors.WithStack(ctx.Err())
	}
	return nil
}

// run starts the forwarding loop for the subscription.
func (s *Subscription[T]) run() {
	defer close(s.quitDone)

	for {
		select {
		case <-s.quit:
			return
		case value := <-s.in:
			select {
			case s.channel <- value:
			case <-s.quit:
				return
			}
		}
	}
}
