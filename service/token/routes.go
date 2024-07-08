package token

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/dawumnam/token-trader/db"
	"github.com/dawumnam/token-trader/service/token/blockchain"
	"github.com/dawumnam/token-trader/service/user/auth"
	"github.com/dawumnam/token-trader/types"
	"github.com/dawumnam/token-trader/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	userRepo  types.UserRepository
	tokenRepo types.TokenRepository
	txManager *db.TxManager
}

func NewHandler(tokenRepo types.TokenRepository, userRepo types.UserRepository, txManager *db.TxManager) *Handler {
	return &Handler{tokenRepo: tokenRepo, txManager: txManager, userRepo: userRepo}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/token/issue", auth.WithJWTAuth(h.handleIssueToken, h.userRepo)).Methods("POST")
	router.HandleFunc("/token/balance/{tokenId}", auth.WithJWTAuth(h.handleGetBalance, h.userRepo)).Methods("GET")
	router.HandleFunc("/token/list", auth.WithJWTAuth(h.handleListTokens, h.userRepo)).Methods("GET")
}

func (h *Handler) handleIssueToken(w http.ResponseWriter, r *http.Request) {
	var payload types.IssueTokenPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	userID := r.Context().Value("userID").(int)
	initialSupply, ok := new(big.Int).SetString(payload.InitialSupply, 10)
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid initial supply"))
		return
	}

	var newToken *types.Token
	err := h.txManager.RunInTransaction(r.Context(), func(tx *sql.Tx) error {
		newToken = &types.Token{
			Name:            payload.Name,
			Symbol:          payload.Symbol,
			ContractAddress: uuid.New().String(),
			OwnerID:         uint(userID),
		}

		tokenManager, err := blockchain.NewTokenManager()
		if err != nil {
			return err
		}

		_, err = tokenManager.DeployToken(
			types.IssueTokenPayload{
				Name:          payload.Name,
				Symbol:        payload.Symbol,
				InitialSupply: payload.InitialSupply,
			},
		)
		if err != nil {
			return err
		}

		err = h.tokenRepo.CreateToken(tx, newToken)
		if err != nil {
			return err
		}

		err = h.tokenRepo.UpdateTokenBalance(tx, uint(userID), newToken.ID, initialSupply)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to issue token: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, newToken)
}

func (h *Handler) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenID, err := strconv.ParseInt(vars["tokenId"], 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid token ID"))
		return
	}

	userID := r.Context().Value("userID").(int)

	var balance *big.Int
	err = h.txManager.RunInTransaction(r.Context(), func(tx *sql.Tx) error {
		var err error
		balance, err = h.tokenRepo.GetTokenBalance(tx, uint(userID), uint(tokenID))
		return err
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get balance: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"balance": balance.String()})
}

func (h *Handler) handleListTokens(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)

	var tokens []*types.Token
	err := h.txManager.RunInTransaction(r.Context(), func(tx *sql.Tx) error {
		var err error
		tokens, err = h.tokenRepo.GetTokensByOwner(tx, uint(userID))
		return err
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to list tokens: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, tokens)
}
