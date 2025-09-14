package models

import "time"

type OrderBookLevel struct {
	Price         float64   `json:"price"`
	Quantity      int       `json:"quantity"`
	Timestamp     time.Time `json:"timestamp"`
	QueuePosition int       `json:"queue_position"`
}

type OrderBookResponse struct {
	Symbol          string           `json:"symbol"`
	Timestamp       time.Time        `json:"timestamp"`
	Bids            []OrderBookLevel `json:"bids"`
	Asks            []OrderBookLevel `json:"asks"`
	Spread          *float64         `json:"spread,omitempty"`
	TotalBidOrders  int              `json:"total_bid_orders"`
	TotalAskOrders  int              `json:"total_ask_orders"`
}