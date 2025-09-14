package database

import "order-matching-engine/models"

func SaveTrade(trade *models.Trade) error {
	query := `INSERT INTO trades (id, symbol, buy_order_id, sell_order_id, price, quantity, executed_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	_, err := DB.Exec(query, trade.ID, trade.Symbol, trade.BuyOrderID, trade.SellOrderID,
		trade.Price, trade.Quantity, trade.ExecutedAt)
	return err
}

func GetTradesBySymbol(symbol string) ([]*models.Trade, error) {
	query := `SELECT id, symbol, buy_order_id, sell_order_id, price, quantity, executed_at 
			  FROM trades WHERE symbol = ? ORDER BY executed_at DESC`
	
	rows, err := DB.Query(query, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []*models.Trade
	for rows.Next() {
		trade := &models.Trade{}
		err := rows.Scan(&trade.ID, &trade.Symbol, &trade.BuyOrderID, &trade.SellOrderID,
			&trade.Price, &trade.Quantity, &trade.ExecutedAt)
		if err != nil {
			return nil, err
		}
		trades = append(trades, trade)
	}
	
	// Check for errors that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return trades, nil
}

func GetAllTrades() ([]*models.Trade, error) {
	query := `SELECT id, symbol, buy_order_id, sell_order_id, price, quantity, executed_at 
			  FROM trades ORDER BY executed_at DESC`
	
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []*models.Trade
	for rows.Next() {
		trade := &models.Trade{}
		err := rows.Scan(&trade.ID, &trade.Symbol, &trade.BuyOrderID, &trade.SellOrderID,
			&trade.Price, &trade.Quantity, &trade.ExecutedAt)
		if err != nil {
			return nil, err
		}
		trades = append(trades, trade)
	}
	
	// Check for errors that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return trades, nil
}