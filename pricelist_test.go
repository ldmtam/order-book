package orderbook

import (
	"context"
	"testing"

	pool "github.com/jolestar/go-commons-pool"
	"github.com/stretchr/testify/assert"
)

type action struct {
	typ string
	val int
}

func TestAscend(t *testing.T) {
	tests := []struct {
		actions []action
		out     []int
	}{
		{
			[]action{{"get", 3}, {"get", 5}, {"get", 10}},
			[]int{3, 5, 10},
		},
		{
			[]action{{"get", 3}, {"get", 5}, {"get", 10}, {"del", 5}},
			[]int{3, 10},
		},
		{
			[]action{{"get", 203}, {"get", 293}, {"get", 1}, {"get", 4}},
			[]int{1, 4, 203, 293},
		},
		{
			[]action{{"get", 203}, {"get", 293}, {"get", 1}, {"get", 4}, {"del", 203}},
			[]int{1, 4, 293},
		},
		{
			[]action{{"get", 100}, {"del", 100}, {"del", 200}},
			[]int{},
		},
	}

	for _, test := range tests {
		priceList := initPriceList(t)
		processAction(t, priceList, test.actions...)

		result := make([]int, 0, priceList.l.GetNodeCount())
		priceList.Ascend(func(item *PriceItem) bool {
			result = append(result, item.Price)
			return true
		})

		assert.Equal(t, test.out, result)
	}
}

func TestDescend(t *testing.T) {
	tests := []struct {
		actions []action
		out     []int
	}{
		{
			[]action{{"get", 3}, {"get", 5}, {"get", 10}},
			[]int{10, 5, 3},
		},
		{
			[]action{{"get", 3}, {"get", 5}, {"get", 10}, {"del", 5}},
			[]int{10, 3},
		},
		{
			[]action{{"get", 203}, {"get", 293}, {"get", 1}, {"get", 4}},
			[]int{293, 203, 4, 1},
		},
		{
			[]action{{"get", 203}, {"get", 293}, {"get", 1}, {"get", 4}, {"del", 203}},
			[]int{293, 4, 1},
		},
		{
			[]action{{"get", 100}, {"del", 100}, {"del", 200}},
			[]int{},
		},
	}

	for _, test := range tests {
		priceList := initPriceList(t)
		processAction(t, priceList, test.actions...)

		result := make([]int, 0, priceList.l.GetNodeCount())
		priceList.Descend(func(item *PriceItem) bool {
			result = append(result, item.Price)
			return true
		})

		assert.Equal(t, test.out, result)
	}
}

func initPriceList(t *testing.T) *PriceList {
	t.Helper()

	ctx := context.WithValue(context.Background(), "order_queue_size", uint64(1))

	cfg := pool.NewDefaultPoolConfig()
	cfg.MaxTotal = 10
	cfg.MaxIdle = -1
	cfg.MinIdle = -1

	orderQueuePool := pool.NewObjectPool(ctx, &OrderQueueObjectFactory{}, cfg)
	pool.Prefill(ctx, orderQueuePool, 10)

	return NewPriceList(orderQueuePool)
}

func processAction(t *testing.T, priceList *PriceList, actions ...action) {
	t.Helper()

	for _, action := range actions {
		switch action.typ {
		case "get":
			priceList.GetItem(action.val)
		case "del":
			priceList.Delete(action.val)
		}
	}
}
