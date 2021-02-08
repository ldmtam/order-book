package orderbook

import (
	"testing"
	"time"

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
