package utils

import "errors"

// Define typed errors for better error handling
var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderAlreadyFilled = errors.New("order already filled")
	ErrOrderCancelled    = errors.New("order already cancelled")
	ErrInvalidOrderStatus = errors.New("invalid order status for operation")
)