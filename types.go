package orderbook

import (
	"fmt"
)

// Side represents order side (buy or sell)
type Side byte

var (
	// Buy represents buy order
	Buy Side = 'B'
	// Sell represents sell order
	Sell Side = 'S'
)

func (s Side) String() string {
	if s == Buy {
		return "Buy"
	}
	return "Sell"
}

// Execution represents the result of a match
type Execution struct {
	BuyOrderID  string
	SellOrderID string
	Quantity    int
	Price       int
	Timestamp   int64
}

func (ex Execution) String() string {
	return fmt.Sprintf("(BuyOrderID: %v, SellOrderID: %v, Quantity: %v, Price: %v, Timestamp: %d)",
		ex.BuyOrderID,
		ex.SellOrderID,
		ex.Quantity,
		ex.Price,
		ex.Timestamp)
}

// Order represents an order
type Order struct {
	Quantity  int
	Price     int
	Timestamp int64
	Side      Side
	ID        string
}

func (o Order) String() string {
	return fmt.Sprintf("ID: %v, Side: %v, Quantity: %v, Price: %v, Timestamp: %d",
		o.ID,
		o.Side,
		o.Quantity,
		o.Price,
		o.Timestamp)
}
