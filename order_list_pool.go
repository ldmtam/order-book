package orderbook

import (
	"context"

	pool "github.com/jolestar/go-commons-pool"
)

// OrderListObjectFactory ...
type OrderListObjectFactory struct{}

// MakeObject ...
func (of *OrderListObjectFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	return &pool.PooledObject{
		Object: NewRingBuffer(_OrderLimit),
	}, nil
}

// DestroyObject ...
func (of *OrderListObjectFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}

// ValidateObject ...
func (of *OrderListObjectFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	return true
}

// ActivateObject ...
func (of *OrderListObjectFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	object.Object.(*RingBuffer).Reset()
	return nil
}

// PassivateObject ...
func (of *OrderListObjectFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}
