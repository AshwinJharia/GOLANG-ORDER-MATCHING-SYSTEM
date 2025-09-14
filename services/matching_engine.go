package services

import (
	"errors"
	"fmt"
	"order-matching-engine/database"
	"order-matching-engine/models"
	"order-matching-engine/utils"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MatchingEngine struct {
	orderBooks map[string]*OrderBook
	mu         sync.RWMutex
}

func NewMatchingEngine() *MatchingEngine {
	return &MatchingEngine{
		orderBooks: make(map[string]*OrderBook),
	}
}

func (me *MatchingEngine) ProcessOrder(order *models.Order) ([]*models.Trade, error) {
	if order == nil {
		return nil, errors.New("order cannot be nil")
	}

	me.mu.Lock()
	defer me.mu.Unlock()

	var trades []*models.Trade
	var updatedOrders []*models.Order

	book := me.getOrderBook(order.Symbol)

	if order.Side == "buy" {
		trades, updatedOrders = me.matchBuyOrder(order, book)
	} else {
		trades, updatedOrders = me.matchSellOrder(order, book)
	}

	// Execute all database operations in a single transaction
	if err := database.ExecuteOrderMatching(order, trades, updatedOrders); err != nil {
		return nil, fmt.Errorf("failed to execute order matching transaction: %w", err)
	}

	return trades, nil
}

func (me *MatchingEngine) getOrderBook(symbol string) *OrderBook {
	if book, exists := me.orderBooks[symbol]; exists {
		return book
	}
	book := NewOrderBook(symbol)
	me.orderBooks[symbol] = book
	return book
}

func (me *MatchingEngine) matchBuyOrder(buyOrder *models.Order, book *OrderBook) ([]*models.Trade, []*models.Order) {
	var trades []*models.Trade
	var updatedOrders []*models.Order

	for i := 0; i < len(book.SellOrders) && buyOrder.RemainingQuantity > 0; {
		sellOrder := book.SellOrders[i]

		if !me.canMatch(buyOrder, sellOrder) {
			break
		}

		trade := me.executeTrade(buyOrder, sellOrder)
		trades = append(trades, trade)

		// Update order statuses
		me.updateOrderStatus(buyOrder)
		me.updateOrderStatus(sellOrder)

		// Track updated orders for transaction
		updatedOrders = append(updatedOrders, sellOrder)

		// Remove fully filled orders
		if sellOrder.RemainingQuantity == 0 {
			book.SellOrders = append(book.SellOrders[:i], book.SellOrders[i+1:]...)
		} else {
			i++
		}
	}

	// Handle market orders with no liquidity
	if buyOrder.Type == "market" && buyOrder.RemainingQuantity > 0 {
		buyOrder.Status = "cancelled"
	} else if buyOrder.RemainingQuantity > 0 && buyOrder.Type == "limit" {
		book.AddOrder(buyOrder)
	}

	return trades, updatedOrders
}

func (me *MatchingEngine) matchSellOrder(sellOrder *models.Order, book *OrderBook) ([]*models.Trade, []*models.Order) {
	var trades []*models.Trade
	var updatedOrders []*models.Order

	for i := 0; i < len(book.BuyOrders) && sellOrder.RemainingQuantity > 0; {
		buyOrder := book.BuyOrders[i]

		if !me.canMatch(buyOrder, sellOrder) {
			break
		}

		trade := me.executeTrade(buyOrder, sellOrder)
		trades = append(trades, trade)

		// Update order statuses
		me.updateOrderStatus(buyOrder)
		me.updateOrderStatus(sellOrder)

		// Track updated orders for transaction
		updatedOrders = append(updatedOrders, buyOrder)

		// Remove fully filled orders
		if buyOrder.RemainingQuantity == 0 {
			book.BuyOrders = append(book.BuyOrders[:i], book.BuyOrders[i+1:]...)
		} else {
			i++
		}
	}

	// Handle market orders with no liquidity
	if sellOrder.Type == "market" && sellOrder.RemainingQuantity > 0 {
		sellOrder.Status = "cancelled"
	} else if sellOrder.RemainingQuantity > 0 && sellOrder.Type == "limit" {
		book.AddOrder(sellOrder)
	}

	return trades, updatedOrders
}

func (me *MatchingEngine) canMatch(buyOrder, sellOrder *models.Order) bool {
	// Market orders can always match
	if buyOrder.Type == "market" || sellOrder.Type == "market" {
		return true
	}

	// For limit orders, buy price must be >= sell price
	if buyOrder.Price != nil && sellOrder.Price != nil {
		return *buyOrder.Price >= *sellOrder.Price
	}

	return false
}

func (me *MatchingEngine) executeTrade(buyOrder, sellOrder *models.Order) *models.Trade {
	// Determine trade quantity (minimum of remaining quantities)
	quantity := buyOrder.RemainingQuantity
	if sellOrder.RemainingQuantity < quantity {
		quantity = sellOrder.RemainingQuantity
	}

	// Determine trade price (use limit order price, or sell price for market orders)
	var price float64
	if sellOrder.Type == "limit" && sellOrder.Price != nil {
		price = *sellOrder.Price
	} else if buyOrder.Type == "limit" && buyOrder.Price != nil {
		price = *buyOrder.Price
	}

	// Update remaining quantities
	buyOrder.RemainingQuantity -= quantity
	sellOrder.RemainingQuantity -= quantity

	return &models.Trade{
		ID:          uuid.New().String(),
		Symbol:      buyOrder.Symbol,
		BuyOrderID:  buyOrder.ID,
		SellOrderID: sellOrder.ID,
		Price:       price,
		Quantity:    quantity,
		ExecutedAt:  time.Now(),
	}
}

func (me *MatchingEngine) updateOrderStatus(order *models.Order) {
	if order.RemainingQuantity == 0 {
		order.Status = "filled"
	} else if order.RemainingQuantity < order.InitialQuantity {
		order.Status = "partial"
	}
}

func (me *MatchingEngine) GetOrder(orderID string) (*models.Order, error) {
	return database.GetOrderByID(orderID)
}

func (me *MatchingEngine) CancelOrder(orderID string) error {
	order, err := database.GetOrderByID(orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return utils.ErrOrderNotFound
	}

	if order.Status == "filled" || order.Status == "cancelled" {
		return fmt.Errorf("cannot cancel order with status: %s", order.Status)
	}

	// Remove from order book
	if book, exists := me.orderBooks[order.Symbol]; exists {
		book.RemoveOrder(orderID)
	}

	// Update status in database
	order.Status = "cancelled"
	return database.UpdateOrder(order)
}

func (me *MatchingEngine) GetOrderBook(symbol string) *OrderBook {
	me.mu.RLock()
	defer me.mu.RUnlock()

	return me.getOrderBook(symbol)
}

func (me *MatchingEngine) GetAllOrderBooks() map[string]interface{} {
	me.mu.RLock()
	defer me.mu.RUnlock()

	allBooks := make(map[string]interface{})
	for symbol, book := range me.orderBooks {
		bids := book.GetTopBids(10)
		asks := book.GetTopAsks(10)
		
		// Format bids and asks (simplified version)
		var formattedBids []map[string]interface{}
		for _, order := range bids {
			if order.Price != nil {
				formattedBids = append(formattedBids, map[string]interface{}{
					"price":    *order.Price,
					"quantity": order.RemainingQuantity,
				})
			}
		}
		
		var formattedAsks []map[string]interface{}
		for _, order := range asks {
			if order.Price != nil {
				formattedAsks = append(formattedAsks, map[string]interface{}{
					"price":    *order.Price,
					"quantity": order.RemainingQuantity,
				})
			}
		}
		
		allBooks[symbol] = map[string]interface{}{
			"symbol": symbol,
			"bids":   formattedBids,
			"asks":   formattedAsks,
		}
	}
	
	return allBooks
}