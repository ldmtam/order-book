package orderbook

import (
	"context"
	"errors"
	"time"

	"github.com/google/btree"
	pool "github.com/jolestar/go-commons-pool"
)

// MatchEngine represents a match engine for a stock symbol
type MatchEngine struct {
	buyPrices       *btree.BTree
	sellPrices      *btree.BTree
	orderQueuePool  *pool.ObjectPool
	cancelledOrders map[string]struct{}
}

// Less ...
func (a *OrderRing) Less(b btree.Item) bool {
	if a.Side == Buy {
		return !(a.Price >= b.(*OrderRing).Price)
	}
	return !(a.Price <= b.(*OrderRing).Price)
}

// NewMatchEngine returns new match engine
func NewMatchEngine(poolSize, orderQueueSize int) *MatchEngine {
	ctx := context.WithValue(context.Background(), "order_queue_size", uint64(orderQueueSize))

	poolConfig := pool.NewDefaultPoolConfig()
	poolConfig.MaxTotal = poolSize
	poolConfig.MaxIdle = -1
	poolConfig.MinIdle = -1

	orderQueuePool := pool.NewObjectPool(ctx, &OrderQueueObjectFactory{}, poolConfig)
	pool.Prefill(ctx, orderQueuePool, poolSize)

	return &MatchEngine{
		buyPrices:       btree.New(5096),
		sellPrices:      btree.New(5096),
		orderQueuePool:  orderQueuePool,
		cancelledOrders: make(map[string]struct{}, 1000),
	}
}

// CancelOrder cancels order
func (engine *MatchEngine) CancelOrder(orderID string) error {
	engine.cancelledOrders[orderID] = struct{}{}
	return nil
}

// ProcessOrder processes buy or sell order
func (engine *MatchEngine) ProcessOrder(order *Order) ([]*Execution, error) {
	switch order.Side {
	case Buy:
		return engine.processBuyOrder(order)
	case Sell:
		return engine.processSellOrder(order)
	default:
		return nil, errors.New("Order side is not correct")
	}
}

func (engine *MatchEngine) processBuyOrder(buyOrder *Order) ([]*Execution, error) {
	// pre-allocate execution list of size 10
	executions := make([]*Execution, 0, 10)

	engine.sellPrices.Descend(func(i btree.Item) bool {
		orderRing := i.(*OrderRing)
		currSellPrice, orderList := orderRing.Price, orderRing.Orders

		if buyOrder.Price < currSellPrice {
			return false
		}

		for orderList.Len() > 0 && buyOrder.Quantity > 0 {
			sellOrder := orderList.Get()
			if sellOrder == nil {
				break
			}

			if _, ok := engine.cancelledOrders[sellOrder.ID]; ok {
				delete(engine.cancelledOrders, sellOrder.ID)
				continue
			}

			if sellOrder.Quantity >= buyOrder.Quantity {
				sellOrder.Quantity -= buyOrder.Quantity

				executions = append(executions, &Execution{
					BuyOrderID:  buyOrder.ID,
					SellOrderID: sellOrder.ID,
					Quantity:    buyOrder.Quantity,
					Price:       currSellPrice,
					Timestamp:   time.Now().UnixNano(),
				})

				buyOrder.Quantity = 0

				if sellOrder.Quantity > 0 {
					if !orderList.Put(sellOrder) {
						// TODO: don't expect this error could happen,
						//        should we panic ?
						break
					}
				}
			} else {
				buyOrder.Quantity -= sellOrder.Quantity

				executions = append(executions, &Execution{
					BuyOrderID:  buyOrder.ID,
					SellOrderID: sellOrder.ID,
					Quantity:    sellOrder.Quantity,
					Price:       currSellPrice,
					Timestamp:   time.Now().UnixNano(),
				})
			}
		}

		if orderList.Len() == 0 {
			if err := engine.orderQueuePool.ReturnObject(context.Background(), orderList); err != nil {
				panic(err)
			}

			if deleted := engine.sellPrices.Delete(i); deleted == nil {
				panic("item doesn't exist")
			}
		}

		if buyOrder.Quantity == 0 || engine.sellPrices.Len() == 0 {
			return false
		}

		return true
	})

	if buyOrder.Quantity > 0 {
		orderRing := &OrderRing{
			Price: buyOrder.Price,
			Side:  Buy,
		}

		item := engine.buyPrices.Get(orderRing)
		if item != nil {
			orderRing = item.(*OrderRing)
		} else {
			item, err := engine.orderQueuePool.BorrowObject(context.Background())
			if err != nil {
				panic(err)
			}

			orderRing.Orders = item.(*OrderQueue)

			engine.buyPrices.ReplaceOrInsert(orderRing)
		}

		if !orderRing.Orders.Put(buyOrder) {
			panic("buy order queue is full")
		}
	}

	return executions, nil
}

func (engine *MatchEngine) processSellOrder(sellOrder *Order) ([]*Execution, error) {
	// pre-allocate execution list of size 10
	executions := make([]*Execution, 0, 10)

	engine.buyPrices.Descend(func(i btree.Item) bool {
		orderRing := i.(*OrderRing)
		currBuyPrice, orderList := orderRing.Price, orderRing.Orders

		if sellOrder.Price > currBuyPrice {
			return false
		}

		for orderList.Len() > 0 && sellOrder.Quantity > 0 {
			buyOrder := orderList.Get()
			if buyOrder == nil {
				break
			}

			if _, ok := engine.cancelledOrders[buyOrder.ID]; ok {
				delete(engine.cancelledOrders, buyOrder.ID)
				continue
			}

			if buyOrder.Quantity >= sellOrder.Quantity {
				buyOrder.Quantity -= sellOrder.Quantity

				executions = append(executions, &Execution{
					BuyOrderID:  buyOrder.ID,
					SellOrderID: sellOrder.ID,
					Quantity:    sellOrder.Quantity,
					Price:       currBuyPrice,
					Timestamp:   time.Now().UnixNano(),
				})

				sellOrder.Quantity = 0

				if buyOrder.Quantity > 0 {
					if !orderList.Put(buyOrder) {
						// TODO: don't expect this error could happen,
						//		 should we panic?
						break
					}
				}
			} else {
				sellOrder.Quantity -= buyOrder.Quantity

				executions = append(executions, &Execution{
					BuyOrderID:  buyOrder.ID,
					SellOrderID: sellOrder.ID,
					Quantity:    buyOrder.Quantity,
					Price:       currBuyPrice,
					Timestamp:   time.Now().UnixNano(),
				})
			}
		}

		if orderList.Len() == 0 {
			if err := engine.orderQueuePool.ReturnObject(context.Background(), orderList); err != nil {
				panic(err)
			}

			if deleted := engine.buyPrices.Delete(i); deleted == nil {
				panic("item doesn't exist")
			}
		}

		if sellOrder.Quantity == 0 || engine.buyPrices.Len() == 0 {
			return false
		}

		return true
	})

	if sellOrder.Quantity > 0 {
		orderRing := &OrderRing{
			Price: sellOrder.Price,
			Side:  Sell,
		}

		item := engine.sellPrices.Get(orderRing)
		if item != nil {
			orderRing = item.(*OrderRing)
		} else {
			item, err := engine.orderQueuePool.BorrowObject(context.Background())
			if err != nil {
				panic(err)
			}

			orderRing.Orders = item.(*OrderQueue)

			engine.sellPrices.ReplaceOrInsert(orderRing)
		}

		if !orderRing.Orders.Put(sellOrder) {
			panic("sell order queue is full")
		}
	}

	return executions, nil
}
