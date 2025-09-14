package services

import (
	"order-matching-engine/models"
	"sort"
	"sync"
)

type OrderBook struct {
	Symbol     string
	BuyOrders  []*models.Order
	SellOrders []*models.Order
	mu         sync.RWMutex
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol:     symbol,
		BuyOrders:  make([]*models.Order, 0),
		SellOrders: make([]*models.Order, 0),
	}
}

func (ob *OrderBook) AddOrder(order *models.Order) {
	if order == nil {
		return // Ignore nil orders
	}
	
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if order.Side == "buy" {
		ob.BuyOrders = ob.insertSortedBuy(ob.BuyOrders, order)
	} else {
		ob.SellOrders = ob.insertSortedSell(ob.SellOrders, order)
	}
}

func (ob *OrderBook) RemoveOrder(orderID string) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	// Remove from buy orders
	for i, order := range ob.BuyOrders {
		if order.ID == orderID {
			ob.BuyOrders = append(ob.BuyOrders[:i], ob.BuyOrders[i+1:]...)
			return
		}
	}

	// Remove from sell orders
	for i, order := range ob.SellOrders {
		if order.ID == orderID {
			ob.SellOrders = append(ob.SellOrders[:i], ob.SellOrders[i+1:]...)
			return
		}
	}
}

func (ob *OrderBook) sortBuyOrders() {
	sort.Slice(ob.BuyOrders, func(i, j int) bool {
		if ob.BuyOrders[i].Price == nil || ob.BuyOrders[j].Price == nil {
			return false
		}
		if *ob.BuyOrders[i].Price == *ob.BuyOrders[j].Price {
			return ob.BuyOrders[i].CreatedAt.Before(ob.BuyOrders[j].CreatedAt)
		}
		return *ob.BuyOrders[i].Price > *ob.BuyOrders[j].Price
	})
}

func (ob *OrderBook) sortSellOrders() {
	sort.Slice(ob.SellOrders, func(i, j int) bool {
		if ob.SellOrders[i].Price == nil || ob.SellOrders[j].Price == nil {
			return false
		}
		if *ob.SellOrders[i].Price == *ob.SellOrders[j].Price {
			return ob.SellOrders[i].CreatedAt.Before(ob.SellOrders[j].CreatedAt)
		}
		return *ob.SellOrders[i].Price < *ob.SellOrders[j].Price
	})
}

func (ob *OrderBook) GetTopBids(limit int) []*models.Order {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if len(ob.BuyOrders) < limit {
		limit = len(ob.BuyOrders)
	}
	return ob.BuyOrders[:limit]
}

func (ob *OrderBook) GetTopAsks(limit int) []*models.Order {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if len(ob.SellOrders) < limit {
		limit = len(ob.SellOrders)
	}
	return ob.SellOrders[:limit]
}

// Optimized insertion functions using binary search
func (ob *OrderBook) insertSortedBuy(orders []*models.Order, newOrder *models.Order) []*models.Order {
	if len(orders) == 0 {
		return []*models.Order{newOrder}
	}

	// Binary search for insertion point (buy orders: highest price first)
	left, right := 0, len(orders)
	for left < right {
		mid := (left + right) / 2
		if ob.compareBuyOrders(newOrder, orders[mid]) {
			right = mid
		} else {
			left = mid + 1
		}
	}

	// Insert at the found position
	orders = append(orders, nil)
	copy(orders[left+1:], orders[left:])
	orders[left] = newOrder
	return orders
}

func (ob *OrderBook) insertSortedSell(orders []*models.Order, newOrder *models.Order) []*models.Order {
	if len(orders) == 0 {
		return []*models.Order{newOrder}
	}

	// Binary search for insertion point (sell orders: lowest price first)
	left, right := 0, len(orders)
	for left < right {
		mid := (left + right) / 2
		if ob.compareSellOrders(newOrder, orders[mid]) {
			right = mid
		} else {
			left = mid + 1
		}
	}

	// Insert at the found position
	orders = append(orders, nil)
	copy(orders[left+1:], orders[left:])
	orders[left] = newOrder
	return orders
}

// Comparison functions for ordering
func (ob *OrderBook) compareBuyOrders(a, b *models.Order) bool {
	if a.Price == nil || b.Price == nil {
		return false
	}
	if *a.Price == *b.Price {
		return a.CreatedAt.Before(b.CreatedAt) // FIFO for same price
	}
	return *a.Price > *b.Price // Higher price first
}

func (ob *OrderBook) compareSellOrders(a, b *models.Order) bool {
	if a.Price == nil || b.Price == nil {
		return false
	}
	if *a.Price == *b.Price {
		return a.CreatedAt.Before(b.CreatedAt) // FIFO for same price
	}
	return *a.Price < *b.Price // Lower price first
}