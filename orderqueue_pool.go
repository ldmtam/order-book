package orderbook

import (
	"context"

	pool "github.com/jolestar/go-commons-pool"
)

// OrderQueueObjectFactory ...
type OrderQueueObjectFactory struct{}

// MakeObject ...
func (of *OrderQueueObjectFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	return &pool.PooledObject{
		Object: NewRingBuffer(_OrderLimit),
	}, nil
}

// DestroyObject ...
func (of *OrderQueueObjectFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}

// ValidateObject ...
func (of *OrderQueueObjectFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	return true
}

// ActivateObject ...
func (of *OrderQueueObjectFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	object.Object.(*RingBuffer).Reset()
	return nil
}

// PassivateObject ...
func (of *OrderQueueObjectFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}
