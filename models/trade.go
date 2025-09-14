package models

import "time"

type Trade struct {
	ID          string    `json:"id" db:"id"`
	Symbol      string    `json:"symbol" db:"symbol"`
	BuyOrderID  string    `json:"buy_order_id" db:"buy_order_id"`
	SellOrderID string    `json:"sell_order_id" db:"sell_order_id"`
	Price       float64   `json:"price" db:"price"`
	Quantity    int       `json:"quantity" db:"quantity"`
	ExecutedAt  time.Time `json:"executed_at" db:"executed_at"`
}