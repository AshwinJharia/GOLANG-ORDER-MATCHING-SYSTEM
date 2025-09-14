package handlers

import (
	"net/http"
	"order-matching-engine/database"
	"order-matching-engine/utils"
)

type TradeHandler struct{}

func NewTradeHandler() *TradeHandler {
	return &TradeHandler{}
}

func (h *TradeHandler) GetTrades(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	
	var trades interface{}
	var err error
	
	if symbol == "" {
		// If no symbol provided, get all trades
		trades, err = database.GetAllTrades()
	} else {
		// Get trades for specific symbol
		trades, err = database.GetTradesBySymbol(symbol)
	}
	
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to get trades")
		return
	}

	utils.WriteSuccess(w, trades)
}