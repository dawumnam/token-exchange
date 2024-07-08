package order

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/dawumnam/token-trader/db"
	"github.com/dawumnam/token-trader/service/user/auth"
	"github.com/dawumnam/token-trader/types"
	"github.com/dawumnam/token-trader/utils"
	"github.com/gorilla/mux"
)

type Handler struct {
	orderRepo types.OrderRepository
	tokenRepo types.TokenRepository
	userRepo  types.UserRepository
	txManager *db.TxManager
}

func NewHandler(orderRepo types.OrderRepository, tokenRepo types.TokenRepository, userRepo types.UserRepository, txManager *db.TxManager) *Handler {
	return &Handler{orderRepo: orderRepo, tokenRepo: tokenRepo, userRepo: userRepo, txManager: txManager}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/order/place", auth.WithJWTAuth(h.handlePlaceOrder, h.userRepo)).Methods("POST")
	router.HandleFunc("/order/list/{tokenId}", auth.WithJWTAuth(h.handleListOrders, h.userRepo)).Methods("GET")
	router.HandleFunc("/order/execute", auth.WithJWTAuth(h.handleExecuteOrder, h.userRepo)).Methods("POST")
	router.HandleFunc("/order/cancel/{orderId}", auth.WithJWTAuth(h.handleCancelOrder, h.userRepo)).Methods("POST")
}

func (h *Handler) handlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	var payload types.PlaceOrderPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	userID := r.Context().Value("userID").(int)
	amount, ok := new(big.Int).SetString(payload.Amount, 10)
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid amount"))
		return
	}
	price, ok := new(big.Int).SetString(payload.Price, 10)
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid price"))
		return
	}

	var newOrder *types.Order
	err := h.txManager.RunInTransaction(r.Context(), func(tx *sql.Tx) error {
		if payload.OrderType == "sell" {
			balance, err := h.tokenRepo.GetTokenBalance(tx, uint(userID), payload.TokenID)
			if err != nil {
				return err
			}
			if balance.Cmp(amount) < 0 {
				return fmt.Errorf("insufficient balance")
			}
			err = h.tokenRepo.UpdateTokenBalance(tx, uint(userID), payload.TokenID, new(big.Int).Sub(balance, amount))
			if err != nil {
				return err
			}
		}

		newOrder = &types.Order{
			UserID:    uint(userID),
			TokenID:   payload.TokenID,
			OrderType: payload.OrderType,
			Amount:    amount,
			Price:     price,
			Status:    "open",
		}

		return h.orderRepo.CreateOrder(tx, newOrder)
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to place order: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, newOrder)
}

func (h *Handler) handleListOrders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenID, err := strconv.ParseUint(vars["tokenId"], 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid token ID"))
		return
	}

	orderType := r.URL.Query().Get("type")
	if orderType != "buy" && orderType != "sell" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid order type"))
		return
	}

	var orders []*types.Order
	err = h.txManager.RunInTransaction(r.Context(), func(tx *sql.Tx) error {
		var err error
		orders, err = h.orderRepo.GetOpenOrders(tx, uint(tokenID), orderType)
		return err
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to list orders: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, orders)
}

func (h *Handler) handleExecuteOrder(w http.ResponseWriter, r *http.Request) {
	var payload types.ExecuteOrderPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	buyerID := r.Context().Value("userID").(int)

	err := h.txManager.RunInTransaction(r.Context(), func(tx *sql.Tx) error {
		order, err := h.orderRepo.GetOrderByID(tx, payload.OrderID)
		if err != nil {
			return err
		}

		if order.Status != "open" {
			return fmt.Errorf("order is not open")
		}
		sellerBalance, err := h.tokenRepo.GetTokenBalance(tx, order.UserID, order.TokenID)
		if err != nil {
			return err
		}

		cmp := sellerBalance.Cmp(order.Amount)
		if cmp < 0 {
			return fmt.Errorf("user has insufficient amount of order")
		}

		err = h.tokenRepo.UpdateTokenBalance(tx, order.UserID, order.TokenID, new(big.Int).Sub(sellerBalance, order.Amount))
		if err != nil {
			return err
		}

		buyerBalance, err := h.tokenRepo.GetTokenBalance(tx, uint(buyerID), order.TokenID)
		if err != nil {
			return err
		}

		err = h.tokenRepo.UpdateTokenBalance(tx, uint(buyerID), order.TokenID, new(big.Int).Add(buyerBalance, order.Amount))
		if err != nil {
			return err
		}

		trade := &types.Trade{
			SellerID: order.UserID,
			BuyerID:  uint(buyerID),
			TokenID:  order.TokenID,
			Amount:   order.Amount,
			Price:    order.Price,
		}
		err = h.orderRepo.CreateTrade(tx, trade)
		if err != nil {
			return err
		}

		order.Status = "filled"
		return h.orderRepo.UpdateOrderStatus(tx, order.ID, order.Status)
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to execute order: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Order executed successfully"})
}

func (h *Handler) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.ParseUint(vars["orderId"], 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid order ID"))
		return
	}

	userID := r.Context().Value("userID").(int)

	err = h.txManager.RunInTransaction(r.Context(), func(tx *sql.Tx) error {
		order, err := h.orderRepo.GetOrderByID(tx, uint(orderID))
		if err != nil {
			return err
		}

		if order.UserID != uint(userID) {
			return fmt.Errorf("not authorized to cancel this order")
		}

		if order.Status != "open" {
			return fmt.Errorf("order is not open")
		}

		if order.OrderType == "sell" {
			balance, err := h.tokenRepo.GetTokenBalance(tx, uint(userID), order.TokenID)
			if err != nil {
				return err
			}
			err = h.tokenRepo.UpdateTokenBalance(tx, uint(userID), order.TokenID, new(big.Int).Add(balance, order.Amount))
			if err != nil {
				return err
			}
		}

		return h.orderRepo.UpdateOrderStatus(tx, order.ID, "cancelled")
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to cancel order: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Order cancelled successfully"})
}
