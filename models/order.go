package models

import "time"

type Order struct {
	ID                string    `json:"id" db:"id"`
	Symbol            string    `json:"symbol" db:"symbol"`
	Side              string    `json:"side" db:"side"` // "buy" or "sell"
	Type              string    `json:"type" db:"type"` // "limit" or "market"
	Price             *float64  `json:"price,omitempty" db:"price"`
	InitialQuantity   int       `json:"initial_quantity" db:"initial_quantity"`
	RemainingQuantity int       `json:"remaining_quantity" db:"remaining_quantity"`
	Status            string    `json:"status" db:"status"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

type PlaceOrderRequest struct {
	Symbol   string   `json:"symbol"`
	Side     string   `json:"side"`
	Type     string   `json:"type"`
	Price    *float64 `json:"price,omitempty"`
	Quantity int      `json:"quantity"`
}