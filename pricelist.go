package orderbook

import (
	"context"
	"fmt"

	pool "github.com/jolestar/go-commons-pool"
	"github.com/ldmtam/skiplist"
)

// PriceItem ...
type PriceItem struct {
	Price  int
	Orders *OrderQueue
}

// ItemIterator ...
type ItemIterator func(i *PriceItem) bool

// PriceList ...
type PriceList struct {
	l              skiplist.SkipList
	orderQueuePool *pool.ObjectPool
}

// ExtractKey ...
func (item *PriceItem) ExtractKey() float64 {
	return float64(item.Price)
}

func (item *PriceItem) String() string {
	return fmt.Sprintf("price: %d", item.Price)
}

// NewPriceList ...
func NewPriceList(orderQueuePool *pool.ObjectPool) *PriceList {
	l := skiplist.New()

	return &PriceList{
		l:              l,
		orderQueuePool: orderQueuePool,
	}
}

// GetItem ...
func (pl *PriceList) GetItem(price int) *PriceItem {
	elem, ok := pl.l.FindByKey(float64(price))
	if ok {
		return elem.GetValue().(*PriceItem)
	}

	obj, err := pl.orderQueuePool.BorrowObject(context.Background())
	if err != nil {
		panic("cannot get object from pool")
	}

	item := &PriceItem{
		Price:  price,
		Orders: obj.(*OrderQueue),
	}

	pl.l.Insert(item)
	return item
}

// Delete ...
func (pl *PriceList) Delete(price int) {
	elem, ok := pl.l.FindByKey(float64(price))
	if !ok {
		return
	}
	pl.orderQueuePool.ReturnObject(context.Background(), elem.GetValue().(*PriceItem).Orders)
	pl.l.Delete(elem.GetValue().(*PriceItem))
}

// Ascend ...
func (pl *PriceList) Ascend(iterator ItemIterator) {
	if pl.l.IsEmpty() {
		return
	}
	smallestElem := pl.l.GetSmallestNode()
	for elem := smallestElem; elem != nil; elem = pl.l.Next(elem) {
		item := elem.GetValue().(*PriceItem)
		if !iterator(item) || pl.l.Next(elem) == smallestElem {
			break
		}
	}
}

// Descend ...
func (pl *PriceList) Descend(iterator ItemIterator) {
	if pl.l.IsEmpty() {
		return
	}
	largestElem := pl.l.GetLargestNode()
	for elem := largestElem; elem != nil; elem = pl.l.Prev(elem) {
		item := elem.GetValue().(*PriceItem)
		if !iterator(item) || pl.l.Prev(elem) == largestElem {
			break
		}
	}
}
