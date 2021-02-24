package orderbook

import (
	"math/rand"
	"testing"

	"github.com/spf13/cast"
)

func BenchmarkPutAndGet(b *testing.B) {
	queue := NewOrderQueue(8192)

	orders := make([]*Order, b.N)
	for i := 0; i < len(orders); i++ {
		orders[i] = &Order{
			ID:    cast.ToString(i),
			Price: i + 100,
		}

		if rand.Intn(2) == 0 {
			orders[i].Side = Buy
		} else {
			orders[i].Side = Sell
		}
	}

	putCnt := 0
	getCnt := 0
	resetCnt := 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if orders[i].Side == Buy {
			putCnt++
			ok := queue.Put(orders[i])
			if !ok {
				resetCnt++
				queue.Reset()
			}
		} else {
			getCnt++
			queue.Get()
		}
	}

	b.Logf("b.N: %d, Put: %d, Get: %d, Reset: %d\n", b.N, putCnt, getCnt, resetCnt)
}

func BenchmarkPutAndReset(b *testing.B) {
	queue := NewOrderQueue(8192)

	orders := make([]*Order, b.N)
	for i := 0; i < len(orders); i++ {
		orders[i] = &Order{
			ID:    cast.ToString(i),
			Price: i + 100,
		}

		if rand.Intn(2) == 0 {
			orders[i].Side = Buy
		} else {
			orders[i].Side = Sell
		}
	}

	putCnt := 0
	resetCnt := 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if orders[i].Side == Buy {
			putCnt++
			ok := queue.Put(orders[i])
			if !ok {
				resetCnt++
				queue.Reset()
			}
		} else {
			resetCnt++
			queue.Reset()
		}
	}

	b.Logf("b.N: %d, Put: %d, Reset: %d\n", b.N, putCnt, resetCnt)
}
