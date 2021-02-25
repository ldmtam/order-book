package orderbook

import (
	"math/rand"
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
)

func TestFilledOrder(t *testing.T) {
	engine := NewMatchEngine(512, 16384)

	tests := []struct {
		orders     []*Order
		executions []Execution
	}{
		{
			[]*Order{
				{
					Quantity:  50,
					Price:     10,
					Timestamp: time.Now().UnixNano(),
					Side:      Buy,
					ID:        "001",
				},
				{
					Quantity:  50,
					Price:     5,
					Timestamp: time.Now().UnixNano(),
					Side:      Sell,
					ID:        "002",
				},
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
		execs := make([]Execution, 0)

		for _, order := range test.orders {
			err := engine.ProcessOrder(order)
			assert.Nil(t, err)

		L:
			for {
				select {
				case exec := <-engine.Execution():
					execs = append(execs, exec)
				case <-time.After(10 * time.Millisecond):
					break L
				}
			}
		}

	}
}

func TestPartialOrder(t *testing.T) {
	engine := NewMatchEngine(512, 16384)

	tests := []struct {
		orders     []*Order
		executions []Execution
	}{
		{
			[]*Order{
				{
					Quantity:  25,
					Price:     10,
					Timestamp: time.Now().UnixNano(),
					Side:      Buy,
					ID:        "001",
				},
				{
					Quantity:  50,
					Price:     5,
					Timestamp: time.Now().UnixNano(),
					Side:      Sell,
					ID:        "002",
				},
				{
					Quantity:  5,
					Price:     3,
					Timestamp: time.Now().UnixNano(),
					Side:      Sell,
					ID:        "003",
				},
				{
					Quantity:  10,
					Price:     5,
					Timestamp: time.Now().UnixNano(),
					Side:      Buy,
					ID:        "004",
				},
			},
			[]Execution{
				{
					BuyOrderID:  "001",
					SellOrderID: "002",
					Quantity:    25,
					Price:       10,
				},
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
		execs := make([]Execution, 0)

		for _, order := range test.orders {
			err := engine.ProcessOrder(order)
			assert.Nil(t, err)

		L:
			for {
				select {
				case exec := <-engine.Execution():
					execs = append(execs, exec)
				case <-time.After(10 * time.Millisecond):
					break L
				}
			}
		}

		assertExecutions(t, test.executions, execs)
	}
}

func TestCancelOrder(t *testing.T) {
	engine := NewMatchEngine(512, 16384)

	order1 := &Order{
		ID:        "001",
		Quantity:  10,
		Price:     5,
		Side:      Buy,
		Timestamp: time.Now().UnixNano(),
	}

	order2 := &Order{
		ID:        "002",
		Quantity:  10,
		Price:     3,
		Side:      Sell,
		Timestamp: time.Now().UnixNano(),
	}

	err := engine.ProcessOrder(order1)
	assert.Nil(t, err)

	err = engine.CancelOrder("001")
	assert.Nil(t, err)

	err = engine.ProcessOrder(order2)
	assert.Nil(t, err)

	var executed bool
	select {
	case <-engine.Execution():
		executed = true
	case <-time.After(10 * time.Millisecond):
		break
	}
	assert.Equal(t, false, executed)
}

func benchmarkProcessOrderRandomInsert(n int, b *testing.B) {
	engine := NewMatchEngine(n, 32768)

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

func assertExecutions(t *testing.T, execs1, execs2 []Execution) {
	t.Helper()

	assert.Equal(t, len(execs1), len(execs2))

	for i := 0; i < len(execs1); i++ {
		assert.Equal(t, execs1[i].BuyOrderID, execs2[i].BuyOrderID)
		assert.Equal(t, execs1[i].Price, execs2[i].Price)
		assert.Equal(t, execs1[i].Quantity, execs2[i].Quantity)
		assert.Equal(t, execs1[i].SellOrderID, execs2[i].SellOrderID)
	}
}
