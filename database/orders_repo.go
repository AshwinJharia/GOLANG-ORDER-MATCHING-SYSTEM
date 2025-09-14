package database

import (
	"database/sql"
	"order-matching-engine/models"
)

func SaveOrder(order *models.Order) error {
	query := `INSERT INTO orders (id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := DB.Exec(query, order.ID, order.Symbol, order.Side, order.Type, order.Price, 
		order.InitialQuantity, order.RemainingQuantity, order.Status, order.CreatedAt)
	return err
}

func UpdateOrder(order *models.Order) error {
	query := `UPDATE orders SET remaining_quantity = ?, status = ? WHERE id = ?`
	_, err := DB.Exec(query, order.RemainingQuantity, order.Status, order.ID)
	return err
}

func GetOrderByID(id string) (*models.Order, error) {
	query := `SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at 
			  FROM orders WHERE id = ?`
	
	row := DB.QueryRow(query, id)
	order := &models.Order{}
	
	err := row.Scan(&order.ID, &order.Symbol, &order.Side, &order.Type, &order.Price,
		&order.InitialQuantity, &order.RemainingQuantity, &order.Status, &order.CreatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return order, err
}

func GetOpenOrdersBySymbol(symbol string) ([]*models.Order, error) {
	query := `SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at 
			  FROM orders WHERE symbol = ? AND status IN ('open', 'partial') 
			  ORDER BY side, price, created_at`
	
	rows, err := DB.Query(query, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		err := rows.Scan(&order.ID, &order.Symbol, &order.Side, &order.Type, &order.Price,
			&order.InitialQuantity, &order.RemainingQuantity, &order.Status, &order.CreatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	
	// Check for errors that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return orders, nil
}