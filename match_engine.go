package orderbook

import (
	"errors"
	"time"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/tidwall/btree"
)

// MatchEngine represents a match engine for a stock symbol
type MatchEngine struct {
	buyPrices  *btree.BTree
	sellPrices *btree.BTree
}

// NewMatchEngine returns new match engine
func NewMatchEngine() *MatchEngine {
	return &MatchEngine{
		/*
			When scanning buy tree for selling order, we want price going from highest to lowest, it means highest price will
			be the root of the tree. Because btree requires Less function for comparision, we have to reverse the comparision
			result as below.
		*/
		buyPrices: btree.New(func(a, b interface{}) bool {
			return !(a.(OrderRing).Price > b.(OrderRing).Price)
		}),
		// Same here for scanning sell tree for buying order.
		sellPrices: btree.New(func(a, b interface{}) bool {
			return !(a.(OrderRing).Price < b.(OrderRing).Price)
		}),
	}
}

// ProcessOrder processes buy or sell order
func (engine *MatchEngine) ProcessOrder(order *Order) ([]Execution, error) {
	switch order.Side {
	case Buy:
		return engine.processBuyOrder(order)
	case Sell:
		return engine.processSellOrder(order)
	default:
		return nil, errors.New("Order side is not correct")
	}
}

func (engine *MatchEngine) processBuyOrder(buyOrder *Order) ([]Execution, error) {
	// pre-allocate execution list of size 10
	executions := make([]Execution, 0, 10)

	engine.sellPrices.Descend(nil, func(i interface{}) bool {
		orderRing := i.(OrderRing)
		currSellPrice, orderList := orderRing.Price, orderRing.Orders

		if buyOrder.Price < currSellPrice {
			return false
		}

		for orderList.Len() > 0 {
			item, err := orderList.Poll(10 * time.Microsecond)
			if err != nil {
				break
			}

			sellOrder := item.(Order)

			if sellOrder.Quantity >= buyOrder.Quantity {
				sellOrder.Quantity -= buyOrder.Quantity

				executions = append(executions, Execution{
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

				return false
			}

			buyOrder.Quantity -= sellOrder.Quantity

			executions = append(executions, Execution{
				BuyOrderID:  buyOrder.ID,
				SellOrderID: sellOrder.ID,
				Quantity:    sellOrder.Quantity,
				Price:       currSellPrice,
				Timestamp:   time.Now().UnixNano(),
			})
		}

		return true
	})

	if buyOrder.Quantity > 0 {
		orderRing := OrderRing{
			Price: buyOrder.Price,
		}

		item := engine.buyPrices.Get(orderRing)
		if item != nil {
			orderRing = item.(OrderRing)
		} else {
			orderRing = OrderRing{
				Price:  buyOrder.Price,
				Orders: queue.NewRingBuffer(_OrderLimit),
			}

			engine.buyPrices.Set(orderRing)
		}

		if err := orderRing.Orders.Put(*buyOrder); err != nil {
			return nil, err
		}
	}

	return executions, nil
}

func (engine *MatchEngine) processSellOrder(sellOrder *Order) ([]Execution, error) {
	// pre-allocate execution list of size 10
	executions := make([]Execution, 0, 10)

	engine.buyPrices.Descend(nil, func(i interface{}) bool {
		orderRing := i.(OrderRing)
		currBuyPrice, orderList := orderRing.Price, orderRing.Orders

		if sellOrder.Price > currBuyPrice {
			return false
		}

		for orderList.Len() > 0 {
			item, err := orderList.Poll(10 * time.Microsecond)
			if err != nil {
				break
			}

			buyOrder := item.(Order)

			if buyOrder.Quantity >= sellOrder.Quantity {
				buyOrder.Quantity -= sellOrder.Quantity

				executions = append(executions, Execution{
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

				return false
			}

			sellOrder.Quantity -= buyOrder.Quantity

			executions = append(executions, Execution{
				BuyOrderID:  buyOrder.ID,
				SellOrderID: sellOrder.ID,
				Quantity:    buyOrder.Quantity,
				Price:       currBuyPrice,
				Timestamp:   time.Now().UnixNano(),
			})
		}

		return true
	})

	if sellOrder.Quantity > 0 {
		orderRing := OrderRing{
			Price: sellOrder.Price,
		}

		item := engine.sellPrices.Get(orderRing)
		if item != nil {
			orderRing = item.(OrderRing)
		} else {
			orderRing = OrderRing{
				Price:  sellOrder.Price,
				Orders: queue.NewRingBuffer(_OrderLimit),
			}

			engine.sellPrices.Set(orderRing)
		}

		if err := orderRing.Orders.Put(*sellOrder); err != nil {
			return nil, err
		}
	}

	return executions, nil
}
