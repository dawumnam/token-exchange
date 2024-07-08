package user

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dawumnam/token-trader/config"
	"github.com/dawumnam/token-trader/db"
	"github.com/dawumnam/token-trader/service/user/auth"
	"github.com/dawumnam/token-trader/types"
	"github.com/dawumnam/token-trader/utils"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type Handler struct {
	repository types.UserRepository
}

func NewHandler(repository types.UserRepository) *Handler {
	return &Handler{repository: repository}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/register", h.HandleRegister).Methods("POST")
	router.HandleFunc("/logout", h.handleLogout).Methods("POST")
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload types.LoginUserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	user, err := h.repository.GetUserByEmail(payload.Email)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("either password or email is invalid"))
		return
	}

	if !auth.ComparePassword([]byte(user.Password), []byte(payload.Password)) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("either password or email is invalid"))
		return
	}

	secret := []byte(config.Envs.JWTSecret)
	token, err := auth.CreateJWT(secret, user.ID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var payload types.RegisterUserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	_, err := h.repository.GetUserByEmail(payload.Email)

	if err == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	u := types.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  hashedPassword,
	}

	err = h.repository.CreateUser(&u)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	secret := []byte(config.Envs.JWTSecret)
	token, err := auth.CreateJWT(secret, u.ID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{"token": token})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	token := auth.GetTokenFromRequest(r)
	if token == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("no token provided"))
		return
	}

	t, err := auth.ValidateToken(token)
	if err != nil || !t.Valid {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid token"))
		return
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("invalid token claims"))
		return
	}

	exp, ok := claims["expiredAt"].(float64)
	if !ok {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("invalid expiration claim"))
		return
	}

	expTime := time.Unix(int64(exp), 0)

	err = db.BlackListToken(token, expTime)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to blacklist token: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Successfully logged out"})
}
