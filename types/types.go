package types

import (
	"database/sql"
	"math/big"
	"time"
)

type UserRepository interface {
	GetUserByEmail(email string) (*User, error)
	GetUserById(id int) (*User, error)
	CreateUser(*User) error
}

type TokenRepository interface {
	CreateToken(tx *sql.Tx, token *Token) error
	GetTokenByID(tx *sql.Tx, id uint) (*Token, error)
	GetTokensByOwner(tx *sql.Tx, ownerID uint) ([]*Token, error)
	UpdateTokenBalance(tx *sql.Tx, userID, tokenID uint, amount *big.Int) error
	GetTokenBalance(tx *sql.Tx, userID, tokenID uint) (*big.Int, error)
}

type OrderRepository interface {
	CreateOrder(tx *sql.Tx, order *Order) error
	GetOrderByID(tx *sql.Tx, id uint) (*Order, error)
	GetOpenOrders(tx *sql.Tx, tokenID uint, orderType string) ([]*Order, error)
	UpdateOrderStatus(tx *sql.Tx, orderID uint, status string) error
	CreateTrade(tx *sql.Tx, trade *Trade) error
	GetUserTrades(tx *sql.Tx, userID uint) ([]*Trade, error)
}

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

type Token struct {
	ID              uint      `json:"id"`
	ContractAddress string    `json:"contractAddress"`
	Name            string    `json:"name"`
	Symbol          string    `json:"symbol"`
	OwnerID         uint      `json:"ownerId"`
	CreatedAt       time.Time `json:"createdAt"`
}

type Balance struct {
	ID      uint     `json:"id"`
	UserID  uint     `json:"userId"`
	TokenID uint     `json:"tokenId"`
	Amount  *big.Int `json:"amount"`
}

type Order struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	TokenID   uint      `json:"tokenId"`
	OrderType string    `json:"orderType"` // "buy" or "sell"
	Amount    *big.Int  `json:"amount"`
	Price     *big.Int  `json:"price"`
	Status    string    `json:"status"` // "open", "filled", or "cancelled"
	CreatedAt time.Time `json:"createdAt"`
}

// Trade represents a completed trade between two users
type Trade struct {
	ID        uint      `json:"id"`
	SellerID  uint      `json:"sellerId"`
	BuyerID   uint      `json:"buyerId"`
	TokenID   uint      `json:"tokenId"`
	Amount    *big.Int  `json:"amount"`
	Price     *big.Int  `json:"price"`
	CreatedAt time.Time `json:"createdAt"`
}

type RegisterUserPayload struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=10,max=30"`
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type IssueTokenPayload struct {
	Name          string `json:"name" validate:"required"`
	Symbol        string `json:"symbol" validate:"required,max=10"`
	InitialSupply string `json:"initialSupply" validate:"required"`
}

type PlaceOrderPayload struct {
	TokenID   uint   `json:"tokenId" validate:"required"`
	OrderType string `json:"orderType" validate:"required,oneof=buy sell"`
	Amount    string `json:"amount" validate:"required"`
	Price     string `json:"price" validate:"required"`
}

type ExecuteOrderPayload struct {
	OrderID uint `json:"orderId" validate:"required"`
}

type GetOpenOrdersPayload struct {
	TokenID uint `json:"tokenId" validate:"required"`
}

type GetUserOrdersPayload struct {
	UserID uint `json:"userId" validate:"required"`
}

type GetUserTradesPayload struct {
	UserID uint `json:"userId" validate:"required"`
}

type GetTokenBalancePayload struct {
	UserID  uint `json:"userId" validate:"required"`
	TokenID uint `json:"tokenId" validate:"required"`
}
