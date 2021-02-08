package orderbook

import (
	"math/rand"
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
)

func TestFilledOrder(t *testing.T) {
	engine := NewMatchEngine()

	tests := []struct {
		order      *Order
		executions []Execution
	}{
		{
			&Order{
				Quantity:  50,
				Price:     10,
				Timestamp: time.Now().UnixNano(),
				Side:      Buy,
				ID:        "001",
			},
			[]Execution{},
		},
		{
			&Order{
				Quantity:  50,
				Price:     5,
				Timestamp: time.Now().UnixNano(),
				Side:      Sell,
				ID:        "002",
			},
			[]Execution{
				{
					BuyOrderID:  "001",
					SellOrderID: "002",
					Quantity:    50,
					Price:       10,
				},
			},
		},
	}

	for _, test := range tests {
		expectedExecutions, err := engine.ProcessOrder(test.order)
		assert.Nil(t, err)

		t.Log(expectedExecutions)

		for idx, expectedExecution := range expectedExecutions {
			assert.EqualValues(t, expectedExecution.BuyOrderID, test.executions[idx].BuyOrderID)
			assert.EqualValues(t, expectedExecution.SellOrderID, test.executions[idx].SellOrderID)
			assert.EqualValues(t, expectedExecution.Quantity, test.executions[idx].Quantity)
			assert.EqualValues(t, expectedExecution.Price, test.executions[idx].Price)
		}
	}
}

func TestPartialOrder(t *testing.T) {
	engine := NewMatchEngine()

	tests := []struct {
		order      *Order
		executions []Execution
	}{
		{
			&Order{
				Quantity:  25,
				Price:     10,
				Timestamp: time.Now().UnixNano(),
				Side:      Buy,
				ID:        "001",
			},
			[]Execution{},
		},
		{
			&Order{
				Quantity:  50,
				Price:     5,
				Timestamp: time.Now().UnixNano(),
				Side:      Sell,
				ID:        "002",
			},
			[]Execution{
				{
					BuyOrderID:  "001",
					SellOrderID: "002",
					Quantity:    25,
					Price:       10,
				},
			},
		},
		{
			&Order{
				Quantity:  5,
				Price:     3,
				Timestamp: time.Now().UnixNano(),
				Side:      Sell,
				ID:        "003",
			},
			[]Execution{},
		},
		{
			&Order{
				Quantity:  10,
				Price:     5,
				Timestamp: time.Now().UnixNano(),
				Side:      Buy,
				ID:        "004",
			},
			[]Execution{
				{
					BuyOrderID:  "004",
					SellOrderID: "003",
					Quantity:    5,
					Price:       3,
				},
				{
					BuyOrderID:  "004",
					SellOrderID: "002",
					Quantity:    5,
					Price:       5,
				},
			},
		},
	}

	for _, test := range tests {
		expectedExecutions, err := engine.ProcessOrder(test.order)
		assert.Nil(t, err)

		t.Log(expectedExecutions)

		for idx, expectedExecution := range expectedExecutions {
			assert.EqualValues(t, expectedExecution.BuyOrderID, test.executions[idx].BuyOrderID)
			assert.EqualValues(t, expectedExecution.SellOrderID, test.executions[idx].SellOrderID)
			assert.EqualValues(t, expectedExecution.Quantity, test.executions[idx].Quantity)
			assert.EqualValues(t, expectedExecution.Price, test.executions[idx].Price)
		}
	}
}

func benchmarkProcessOrderRandomInsert(n int, b *testing.B) {
	engine := NewMatchEngine()

	prices := make([]int, n)
	for i := range prices {
		prices[i] = rand.Intn(n) + 1
	}

	orders := make([]*Order, 0, b.N)
	for i := 0; i < b.N; i++ {
		orders = append(orders, &Order{})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		price := prices[rand.Intn(n)]

		order := orders[i]
		order.ID = cast.ToString(i)
		order.Price = price
		order.Quantity = rand.Int()

		if price < n/2 {
			order.Side = Buy
		} else {
			order.Side = Sell
		}

		engine.ProcessOrder(order)
	}
}

func BenchmarkProcessOrder1kLevels(b *testing.B) {
	benchmarkProcessOrderRandomInsert(1000, b)
}

func BenchmarkProcessOrder5kLevels(b *testing.B) {
	benchmarkProcessOrderRandomInsert(5000, b)
}

func BenchmarkProcessOrder10kLevels(b *testing.B) {
	benchmarkProcessOrderRandomInsert(10000, b)
}

func BenchmarkProcessOrder20kLevels(b *testing.B) {
	benchmarkProcessOrderRandomInsert(20000, b)
}
