package database

import (
	"database/sql"
	"fmt"
	"order-matching-engine/models"
)

// ExecuteOrderMatching performs all order matching operations in a single transaction
func ExecuteOrderMatching(order *models.Order, trades []*models.Trade, updatedOrders []*models.Order) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() succeeds

	// Save the new order
	if err := saveOrderTx(tx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Save all trades
	for _, trade := range trades {
		if err := saveTradeTx(tx, trade); err != nil {
			return fmt.Errorf("failed to save trade %s: %w", trade.ID, err)
		}
	}

	// Update all modified orders
	for _, updatedOrder := range updatedOrders {
		if err := updateOrderTx(tx, updatedOrder); err != nil {
			return fmt.Errorf("failed to update order %s: %w", updatedOrder.ID, err)
		}
	}

	return tx.Commit()
}

func saveOrderTx(tx *sql.Tx, order *models.Order) error {
	query := `INSERT INTO orders (id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.Exec(query, order.ID, order.Symbol, order.Side, order.Type, order.Price, 
		order.InitialQuantity, order.RemainingQuantity, order.Status, order.CreatedAt)
	return err
}

func saveTradeTx(tx *sql.Tx, trade *models.Trade) error {
	query := `INSERT INTO trades (id, symbol, buy_order_id, sell_order_id, price, quantity, executed_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.Exec(query, trade.ID, trade.Symbol, trade.BuyOrderID, trade.SellOrderID, 
		trade.Price, trade.Quantity, trade.ExecutedAt)
	return err
}

func updateOrderTx(tx *sql.Tx, order *models.Order) error {
	query := `UPDATE orders SET remaining_quantity = ?, status = ? WHERE id = ?`
	_, err := tx.Exec(query, order.RemainingQuantity, order.Status, order.ID)
	return err
}