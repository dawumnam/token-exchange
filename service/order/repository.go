package order

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/dawumnam/token-trader/types"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(tx *sql.Tx, order *types.Order) error {
	query := `INSERT INTO orders (userID, tokenID, orderType, amount, price, status) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := tx.Exec(query, order.UserID, order.TokenID, order.OrderType, order.Amount.String(), order.Price.String(), order.Status)
	if err != nil {
		return fmt.Errorf("error creating order: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	order.ID = uint(id)
	return nil
}

func (r *OrderRepository) GetOrderByID(tx *sql.Tx, id uint) (*types.Order, error) {
	query := `SELECT id, userID, tokenID, orderType, amount, price, status, createdAt FROM orders WHERE id = ?`
	var order types.Order
	var amountStr, priceStr string
	err := tx.QueryRow(query, id).Scan(
		&order.ID, &order.UserID, &order.TokenID, &order.OrderType, &amountStr, &priceStr, &order.Status, &order.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("error getting order: %w", err)
	}

	order.Amount, _ = new(big.Int).SetString(amountStr, 10)
	order.Price, _ = new(big.Int).SetString(priceStr, 10)

	return &order, nil
}

func (r *OrderRepository) GetOpenOrders(tx *sql.Tx, tokenID uint, orderType string) ([]*types.Order, error) {
	query := `SELECT id, userID, tokenID, orderType, amount, price, status, createdAt 
              FROM orders 
              WHERE tokenID = ? AND orderType = ? AND status = 'open'
              ORDER BY price ASC, createdAt ASC`
	rows, err := tx.Query(query, tokenID, orderType)
	if err != nil {
		return nil, fmt.Errorf("error getting open orders: %w", err)
	}
	defer rows.Close()

	var orders []*types.Order
	for rows.Next() {
		var order types.Order
		var amountStr, priceStr string
		err := rows.Scan(&order.ID, &order.UserID, &order.TokenID, &order.OrderType, &amountStr, &priceStr, &order.Status, &order.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning order: %w", err)
		}
		order.Amount, _ = new(big.Int).SetString(amountStr, 10)
		order.Price, _ = new(big.Int).SetString(priceStr, 10)
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) UpdateOrderStatus(tx *sql.Tx, orderID uint, status string) error {
	query := `UPDATE orders SET status = ? WHERE id = ?`
	_, err := tx.Exec(query, status, orderID)
	if err != nil {
		return fmt.Errorf("error updating order status: %w", err)
	}
	return nil
}

func (r *OrderRepository) CreateTrade(tx *sql.Tx, trade *types.Trade) error {
	query := `INSERT INTO trades (sellerID, buyerID, tokenID, amount, price) VALUES (?, ?, ?, ?, ?)`
	result, err := tx.Exec(query, trade.SellerID, trade.BuyerID, trade.TokenID, trade.Amount.String(), trade.Price.String())
	if err != nil {
		return fmt.Errorf("error creating trade: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	trade.ID = uint(id)
	return nil
}

func (r *OrderRepository) GetUserTrades(tx *sql.Tx, userID uint) ([]*types.Trade, error) {
	query := `SELECT id, sellerID, buyerID, tokenID, amount, price, createdAt 
              FROM trades 
              WHERE sellerID = ? OR buyerID = ?
              ORDER BY createdAt DESC`
	rows, err := tx.Query(query, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user trades: %w", err)
	}
	defer rows.Close()

	var trades []*types.Trade
	for rows.Next() {
		var trade types.Trade
		var amountStr, priceStr string
		err := rows.Scan(&trade.ID, &trade.SellerID, &trade.BuyerID, &trade.TokenID, &amountStr, &priceStr, &trade.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning trade: %w", err)
		}
		trade.Amount, _ = new(big.Int).SetString(amountStr, 10)
		trade.Price, _ = new(big.Int).SetString(priceStr, 10)
		trades = append(trades, &trade)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trades: %w", err)
	}

	return trades, nil
}
