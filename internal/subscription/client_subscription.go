package subscription

import "context"

// ClientSubscription is a subscription that can be used by the client to unsubscribe from the subscription.
type ClientSubscription[T any] struct {
	subscription *Subscription[T]
}

func (c *ClientSubscription[T]) Unsubscribe() {
	c.subscription.Unsubscribe()
}

func (c *ClientSubscription[T]) UnsubscribeWithContext(ctx context.Context) (err error) {
	return c.subscription.UnsubscribeWithContext(ctx)
}

// Err returns the error channel of the subscription.
func (c *ClientSubscription[T]) Err() <-chan error {
	return c.subscription.Err()
}

// Done returns the done channel of the subscription
func (c *ClientSubscription[T]) Done() <-chan struct{} {
	return c.subscription.Done()
}

// IsClosed returns status of the subscription
func (c *ClientSubscription[T]) IsClosed() bool {
	return c.subscription.IsClosed()
}
