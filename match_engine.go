package orderbook

import (
	"context"
	"errors"
	"time"

	pool "github.com/jolestar/go-commons-pool"
)

// MatchEngine represents a match engine for a stock symbol
type MatchEngine struct {
	buyPrices       *PriceList
	sellPrices      *PriceList
	cancelledOrders map[string]struct{}
	executionCh     chan Execution
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
		buyPrices:       NewPriceList(orderQueuePool),
		sellPrices:      NewPriceList(orderQueuePool),
		cancelledOrders: make(map[string]struct{}, 1000),
		executionCh:     make(chan Execution, 1e5),
	}
}

// Execution ...
func (engine *MatchEngine) Execution() <-chan Execution {
	return engine.executionCh
}

// CancelOrder cancels order
func (engine *MatchEngine) CancelOrder(orderID string) error {
	engine.cancelledOrders[orderID] = struct{}{}
	return nil
}

// ProcessOrder processes buy or sell order
func (engine *MatchEngine) ProcessOrder(order *Order) error {
	switch order.Side {
	case Buy:
		return engine.processBuyOrder(order)
	case Sell:
		return engine.processSellOrder(order)
	default:
		return errors.New("Order side is not correct")
	}
}

func (engine *MatchEngine) processBuyOrder(buyOrder *Order) error {
	engine.sellPrices.Ascend(func(item *PriceItem) bool {
		currSellPrice, orderList := item.Price, item.Orders

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

				engine.executionCh <- Execution{
					BuyOrderID:  buyOrder.ID,
					SellOrderID: sellOrder.ID,
					Quantity:    buyOrder.Quantity,
					Price:       currSellPrice,
					Timestamp:   time.Now().UnixNano(),
				}

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

				engine.executionCh <- Execution{
					BuyOrderID:  buyOrder.ID,
					SellOrderID: sellOrder.ID,
					Quantity:    sellOrder.Quantity,
					Price:       currSellPrice,
					Timestamp:   time.Now().UnixNano(),
				}
			}
		}

		if orderList.Len() == 0 {
			engine.sellPrices.Delete(currSellPrice)
		}

		if buyOrder.Quantity == 0 {
			return false
		}

		return true
	})

	if buyOrder.Quantity > 0 {
		item := engine.buyPrices.GetItem(buyOrder.Price)
		if !item.Orders.Put(buyOrder) {
			panic("buy order queue is full")
		}
	}

	return nil
}

func (engine *MatchEngine) processSellOrder(sellOrder *Order) error {
	engine.buyPrices.Descend(func(item *PriceItem) bool {
		currBuyPrice, orderList := item.Price, item.Orders

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

				sellOrder.Quantity = 0

				if buyOrder.Quantity > 0 {
					if !orderList.Put(buyOrder) {
						break
					}
				}
			} else {
				sellOrder.Quantity -= buyOrder.Quantity

				engine.executionCh <- Execution{
					BuyOrderID:  buyOrder.ID,
					SellOrderID: sellOrder.ID,
					Quantity:    sellOrder.Quantity,
					Price:       currBuyPrice,
					Timestamp:   time.Now().UnixNano(),
				}
			}
		}

		if orderList.Len() == 0 {
			engine.buyPrices.Delete(currBuyPrice)
		}

		if sellOrder.Quantity == 0 {
			return false
		}

		return true
	})

	if sellOrder.Quantity > 0 {
		item := engine.sellPrices.GetItem(sellOrder.Price)
		if !item.Orders.Put(sellOrder) {
			panic("sell order queue is full")
		}
	}

	return nil
}
