package token

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/dawumnam/token-trader/types"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) CreateToken(tx *sql.Tx, token *types.Token) error {
	query := `INSERT INTO tokens (contractAddress, name, symbol, ownerID) VALUES (?, ?, ?, ?)`
	result, err := tx.Exec(query, token.ContractAddress, token.Name, token.Symbol, token.OwnerID)
	if err != nil {
		return fmt.Errorf("error creating token: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	token.ID = uint(id)
	return nil
}

func (r *TokenRepository) GetTokenByID(tx *sql.Tx, id uint) (*types.Token, error) {
	query := `SELECT id, contractAddress, name, symbol, ownerID, createdAt FROM tokens WHERE id = ?`
	var token types.Token
	err := tx.QueryRow(query, id).Scan(
		&token.ID, &token.ContractAddress, &token.Name, &token.Symbol, &token.OwnerID, &token.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("error getting token: %w", err)
	}
	return &token, nil
}

func (r *TokenRepository) GetTokensByOwner(tx *sql.Tx, ownerID uint) ([]*types.Token, error) {
	query := `SELECT id, contractAddress, name, symbol, ownerID, createdAt FROM tokens WHERE ownerID = ?`
	rows, err := tx.Query(query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("error getting tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*types.Token
	for rows.Next() {
		var token types.Token
		err := rows.Scan(&token.ID, &token.ContractAddress, &token.Name, &token.Symbol, &token.OwnerID, &token.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning token: %w", err)
		}
		tokens = append(tokens, &token)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tokens: %w", err)
	}

	return tokens, nil
}

func (r *TokenRepository) UpdateTokenBalance(tx *sql.Tx, userID, tokenID uint, amount *big.Int) error {
	query := `INSERT INTO balances (userID, tokenID, amount) VALUES (?, ?, ?)
              ON DUPLICATE KEY UPDATE amount = ?`
	_, err := tx.Exec(query, userID, tokenID, amount.String(), amount.String())
	if err != nil {
		return fmt.Errorf("error updating token balance: %w", err)
	}
	return nil
}

func (r *TokenRepository) GetTokenBalance(tx *sql.Tx, userID, tokenID uint) (*big.Int, error) {
	query := `SELECT amount FROM balances WHERE userID = ? AND tokenID = ?`
	var amountStr string
	err := tx.QueryRow(query, userID, tokenID).Scan(&amountStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return big.NewInt(0), nil
		}
		return nil, fmt.Errorf("error getting token balance: %w", err)
	}

	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return nil, fmt.Errorf("error parsing balance amount")
	}

	return amount, nil
}
