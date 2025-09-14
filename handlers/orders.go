package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"order-matching-engine/models"
	"order-matching-engine/services"
	"order-matching-engine/utils"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type OrderHandler struct {
	engine *services.MatchingEngine
}

func NewOrderHandler(engine *services.MatchingEngine) *OrderHandler {
	return &OrderHandler{engine: engine}
}

func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	// Validate Content-Type
	if !h.validateContentType(w, r) {
		return
	}

	var req models.PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate order request
	if err := h.validateOrderRequest(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create order
	order := &models.Order{
		ID:                uuid.New().String(),
		Symbol:            req.Symbol,
		Side:              req.Side,
		Type:              req.Type,
		Price:             req.Price,
		InitialQuantity:   req.Quantity,
		RemainingQuantity: req.Quantity,
		Status:            "open",
		CreatedAt:         time.Now(),
	}

	// Process order through matching engine
	trades, err := h.engine.ProcessOrder(order)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"order":  order,
		"trades": trades,
	}

	utils.WriteSuccess(w, response)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	if orderID == "" {
		utils.WriteError(w, http.StatusBadRequest, "Order ID required")
		return
	}

	order, err := h.engine.GetOrder(orderID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if order == nil {
		utils.WriteError(w, http.StatusNotFound, "Order not found")
		return
	}

	utils.WriteSuccess(w, order)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	if orderID == "" {
		utils.WriteError(w, http.StatusBadRequest, "Order ID required")
		return
	}

	err := h.engine.CancelOrder(orderID)
	if err != nil {
		if errors.Is(err, utils.ErrOrderNotFound) {
			utils.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			utils.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	utils.WriteSuccess(w, map[string]string{"message": "Order cancelled successfully"})
}

func (h *OrderHandler) GetOrderBook(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	
	if symbol == "" {
		// Return all order books if no symbol specified
		allBooks := h.engine.GetAllOrderBooks()
		utils.WriteSuccess(w, allBooks)
		return
	}

	book := h.engine.GetOrderBook(symbol)

	bids := book.GetTopBids(10)
	asks := book.GetTopAsks(10)
	
	formattedBids := h.formatBidsWithTimestamp(bids)
	formattedAsks := h.formatAsksWithTimestamp(asks)
	
	response := models.OrderBookResponse{
		Symbol:         symbol,
		Timestamp:      time.Now(),
		Bids:           formattedBids,
		Asks:           formattedAsks,
		Spread:         h.calculateSpread(formattedBids, formattedAsks),
		TotalBidOrders: len(bids),
		TotalAskOrders: len(asks),
	}

	utils.WriteSuccess(w, response)
}

func (h *OrderHandler) formatBidsWithTimestamp(orders []*models.Order) []models.OrderBookLevel {
	if len(orders) == 0 {
		return []models.OrderBookLevel{} // Always return empty array, never nil
	}
	
	type priceInfo struct {
		Quantity      int
		EarliestTime  time.Time
		QueuePosition int
	}
	
	priceMap := make(map[float64]*priceInfo)
	
	for i, order := range orders {
		if order.Price != nil {
			if info, exists := priceMap[*order.Price]; exists {
				info.Quantity += order.RemainingQuantity
				if order.CreatedAt.Before(info.EarliestTime) {
					info.EarliestTime = order.CreatedAt
					info.QueuePosition = i + 1
				}
			} else {
				priceMap[*order.Price] = &priceInfo{
					Quantity:      order.RemainingQuantity,
					EarliestTime:  order.CreatedAt,
					QueuePosition: i + 1,
				}
			}
		}
	}
	
	var levels []models.OrderBookLevel
	for price, info := range priceMap {
		levels = append(levels, models.OrderBookLevel{
			Price:         price,
			Quantity:      info.Quantity,
			Timestamp:     info.EarliestTime,
			QueuePosition: info.QueuePosition,
		})
	}
	
	// Sort bids by price DESC (highest first)
	sort.Slice(levels, func(i, j int) bool {
		return levels[i].Price > levels[j].Price
	})
	
	return levels
}

func (h *OrderHandler) formatAsksWithTimestamp(orders []*models.Order) []models.OrderBookLevel {
	if len(orders) == 0 {
		return []models.OrderBookLevel{} // Always return empty array, never nil
	}
	
	type priceInfo struct {
		Quantity      int
		EarliestTime  time.Time
		QueuePosition int
	}
	
	priceMap := make(map[float64]*priceInfo)
	
	for i, order := range orders {
		if order.Price != nil {
			if info, exists := priceMap[*order.Price]; exists {
				info.Quantity += order.RemainingQuantity
				if order.CreatedAt.Before(info.EarliestTime) {
					info.EarliestTime = order.CreatedAt
					info.QueuePosition = i + 1
				}
			} else {
				priceMap[*order.Price] = &priceInfo{
					Quantity:      order.RemainingQuantity,
					EarliestTime:  order.CreatedAt,
					QueuePosition: i + 1,
				}
			}
		}
	}
	
	var levels []models.OrderBookLevel
	for price, info := range priceMap {
		levels = append(levels, models.OrderBookLevel{
			Price:         price,
			Quantity:      info.Quantity,
			Timestamp:     info.EarliestTime,
			QueuePosition: info.QueuePosition,
		})
	}
	
	// Sort asks by price ASC (lowest first)
	sort.Slice(levels, func(i, j int) bool {
		return levels[i].Price < levels[j].Price
	})
	
	return levels
}

func (h *OrderHandler) calculateSpread(bids, asks []models.OrderBookLevel) *float64 {
	if len(bids) == 0 || len(asks) == 0 {
		return nil
	}
	
	// Best bid is highest price (first in sorted bids)
	// Best ask is lowest price (first in sorted asks)
	spread := asks[0].Price - bids[0].Price
	return &spread
}

// Helper functions for validation
func (h *OrderHandler) validateContentType(w http.ResponseWriter, r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		utils.WriteError(w, http.StatusBadRequest, "Content-Type must be application/json")
		return false
	}
	return true
}

func (h *OrderHandler) validateOrderRequest(req *models.PlaceOrderRequest) error {
	// Validate required fields
	if req.Symbol == "" {
		return errors.New("symbol is required")
	}
	if len(req.Symbol) > 50 {
		return errors.New("symbol too long (max 50 characters)")
	}
	if req.Side == "" {
		return errors.New("side is required")
	}
	if req.Type == "" {
		return errors.New("type is required")
	}
	if req.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	// Validate side
	if req.Side != "buy" && req.Side != "sell" {
		return errors.New("side must be 'buy' or 'sell'")
	}

	// Validate type
	if req.Type != "limit" && req.Type != "market" {
		return errors.New("type must be 'limit' or 'market'")
	}

	// Validate price for limit orders
	if req.Type == "limit" {
		if req.Price == nil {
			return errors.New("price required for limit orders")
		}
		if *req.Price <= 0 {
			return errors.New("price must be positive")
		}
	}

	// Market orders should not have price
	if req.Type == "market" && req.Price != nil {
		return errors.New("market orders should not have price")
	}

	return nil
}