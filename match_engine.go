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
	buyPrices     *btree.BTree
	sellPrices    *btree.BTree
	orderListPool *pool.ObjectPool
}

// Less ...
func (a OrderRing) Less(b btree.Item) bool {
	if a.Side == Buy {
		return !(a.Price >= b.(OrderRing).Price)
	}
	return !(a.Price <= b.(OrderRing).Price)
}

// NewMatchEngine returns new match engine
func NewMatchEngine(poolSize int) *MatchEngine {
	ctx := context.Background()

	poolConfig := pool.NewDefaultPoolConfig()
	poolConfig.MaxTotal = poolSize
	poolConfig.MaxIdle = poolSize

	orderListPool := pool.NewObjectPool(ctx, &OrderQueueObjectFactory{}, poolConfig)

	for i := 0; i < poolSize; i++ {
		orderListPool.AddObject(ctx)
	}

	return &MatchEngine{
		/*
			When scanning buy tree for selling order, we want price going from highest to lowest, it means highest price will
			be the root of the tree. Because btree requires Less function for comparision, we have to reverse the comparision
			result as below.
		*/
		buyPrices: btree.New(5096),
		// Same here for scanning sell tree for buying order.
		sellPrices:    btree.New(5096),
		orderListPool: orderListPool,
	}
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
		orderRing := i.(OrderRing)
		currSellPrice, orderList := orderRing.Price, orderRing.Orders

		if buyOrder.Price < currSellPrice {
			return false
		}

		for orderList.Len() > 0 && buyOrder.Quantity > 0 {
			item, err := orderList.Poll(10 * time.Microsecond)
			if err != nil {
				break
			}

			sellOrder := item.(*Order)

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
					if err := orderList.Put(sellOrder); err != nil {
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
			// TODO
			if err := engine.orderListPool.ReturnObject(context.Background(), orderList); err != nil {
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
		orderRing := OrderRing{
			Price: buyOrder.Price,
			Side:  Buy,
		}

		item := engine.buyPrices.Get(orderRing)
		if item != nil {
			orderRing = item.(OrderRing)
		} else {
			item, err := engine.orderListPool.BorrowObject(context.Background())
			if err != nil {
				panic(err)
			}

			orderRing = OrderRing{
				Price:  buyOrder.Price,
				Orders: item.(*RingBuffer),
				Side:   Buy,
			}

			engine.buyPrices.ReplaceOrInsert(orderRing)
		}

		if err := orderRing.Orders.Put(buyOrder); err != nil {
			return nil, err
		}
	}

	return executions, nil
}

func (engine *MatchEngine) processSellOrder(sellOrder *Order) ([]*Execution, error) {
	// pre-allocate execution list of size 10
	executions := make([]*Execution, 0, 10)

	engine.buyPrices.Descend(func(i btree.Item) bool {
		orderRing := i.(OrderRing)
		currBuyPrice, orderList := orderRing.Price, orderRing.Orders

		if sellOrder.Price > currBuyPrice {
			return false
		}

		for orderList.Len() > 0 && sellOrder.Quantity > 0 {
			item, err := orderList.Poll(10 * time.Microsecond)
			if err != nil {
				break
			}

			buyOrder := item.(*Order)

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
					if err := orderList.Put(buyOrder); err != nil {
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
			// TODO
			if err := engine.orderListPool.ReturnObject(context.Background(), orderList); err != nil {
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
		orderRing := OrderRing{
			Price: sellOrder.Price,
			Side:  Sell,
		}

		item := engine.sellPrices.Get(orderRing)
		if item != nil {
			orderRing = item.(OrderRing)
		} else {
			item, err := engine.orderListPool.BorrowObject(context.Background())
			if err != nil {
				panic(err)
			}

			orderRing = OrderRing{
				Price:  sellOrder.Price,
				Orders: item.(*RingBuffer),
				Side:   Sell,
			}

			engine.sellPrices.ReplaceOrInsert(orderRing)
		}

		if err := orderRing.Orders.Put(sellOrder); err != nil {
			return nil, err
		}

	}

	return executions, nil
}
