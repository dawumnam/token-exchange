package order

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dawumnam/token-trader/config"
	"github.com/dawumnam/token-trader/db"
	"github.com/dawumnam/token-trader/service/token"
	"github.com/dawumnam/token-trader/service/user"
	"github.com/dawumnam/token-trader/types"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var testDB *sql.DB
var orderHandler *Handler
var userHandler *user.Handler
var tokenHandler *token.Handler

func TestMain(m *testing.M) {
	cfg := config.Envs
	dbConfig := mysql.Config{
		User:                 cfg.DBUser,
		Passwd:               cfg.DBPassword,
		Addr:                 cfg.DBAddress,
		DBName:               cfg.DBName,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	var err error
	testDB, err = sql.Open("mysql", dbConfig.FormatDSN())
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	db.InitDatabase(testDB)
	db.Init()

	orderRepo := NewOrderRepository(testDB)
	tokenRepo := token.NewTokenRepository(testDB)
	userRepo := user.NewRepository(testDB)
	txManager := db.NewTxManager(testDB)

	orderHandler = NewHandler(orderRepo, tokenRepo, userRepo, txManager)
	userHandler = user.NewHandler(userRepo)
	tokenHandler = token.NewHandler(tokenRepo, userRepo, txManager)

	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

func createRandomUser(t *testing.T) (types.User, string) {
	payload := types.RegisterUserPayload{
		Email:     fmt.Sprintf("user_%s@example.com", uuid.New().String()),
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/register", userHandler.HandleRegister)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Fatalf("Failed to create user: got status %v", status)
	}

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	token := response["token"]
	if token == "" {
		t.Fatalf("No token returned in response")
	}

	return types.User{
		Email:     payload.Email,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
	}, token
}

func createTokenForUser(t *testing.T, token string) types.Token {
	payload := types.IssueTokenPayload{
		Name:          "Test Token",
		Symbol:        "TST",
		InitialSupply: "1000",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/token/issue", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	tokenHandler.RegisterRoutes(router)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Fatalf("Failed to create token: got status %v", status)
	}

	var response types.Token
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	return response
}

func TestHandlePlaceOrder(t *testing.T) {
	_, token := createRandomUser(t)
	createdToken := createTokenForUser(t, token)

	payload := types.PlaceOrderPayload{
		TokenID:   createdToken.ID,
		OrderType: "buy",
		Amount:    "100",
		Price:     "10",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/order/place", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	orderHandler.RegisterRoutes(router)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response types.Order
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.TokenID != createdToken.ID || response.OrderType != "buy" || response.Amount.String() != "100" || response.Price.String() != "10" || response.Status != "open" {
		t.Errorf("Handler returned unexpected body: %+v", response)
	}
}

func TestHandleListOrders(t *testing.T) {
	_, token := createRandomUser(t)
	createdToken := createTokenForUser(t, token)

	for i := 0; i < 2; i++ {
		payload := types.PlaceOrderPayload{
			TokenID:   createdToken.ID,
			OrderType: "buy",
			Amount:    fmt.Sprintf("%d", 100*(i+1)),
			Price:     "10",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/order/place", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		orderHandler.RegisterRoutes(router)
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
		}
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("/order/list/%d?type=buy", createdToken.ID), nil)
	req.Header.Set("Authorization", token)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	orderHandler.RegisterRoutes(router)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var orders []*types.Order
	err := json.Unmarshal(rr.Body.Bytes(), &orders)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("Unexpected number of orders: got %v want %v", len(orders), 2)
	}
}

func TestHandleCancelOrder(t *testing.T) {
	_, token := createRandomUser(t)
	createdToken := createTokenForUser(t, token)

	placePayload := types.PlaceOrderPayload{
		TokenID:   createdToken.ID,
		OrderType: "buy",
		Amount:    "100",
		Price:     "10",
	}

	body, _ := json.Marshal(placePayload)
	req, _ := http.NewRequest("POST", "/order/place", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	orderHandler.RegisterRoutes(router)
	router.ServeHTTP(rr, req)

	var placedOrder types.Order
	json.Unmarshal(rr.Body.Bytes(), &placedOrder)

	req, _ = http.NewRequest("POST", fmt.Sprintf("/order/cancel/%d", placedOrder.ID), nil)
	req.Header.Set("Authorization", token)
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if message, exists := response["message"]; !exists || message != "Order cancelled successfully" {
		t.Errorf("Handler returned unexpected message: got %v want %v", message, "Order cancelled successfully")
	}

}

func TestHandleExecuteOrder(t *testing.T) {
	_, sellerToken := createRandomUser(t)
	createdToken := createTokenForUser(t, sellerToken)

	sellOrderPayload := types.PlaceOrderPayload{
		TokenID:   createdToken.ID,
		OrderType: "sell",
		Amount:    "100",
		Price:     "10",
	}

	body, _ := json.Marshal(sellOrderPayload)
	req, _ := http.NewRequest("POST", "/order/place", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", sellerToken)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	orderHandler.RegisterRoutes(router)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Fatalf("Failed to place sell order: got status %v", status)
	}

	var sellOrder types.Order
	err := json.Unmarshal(rr.Body.Bytes(), &sellOrder)
	if err != nil {
		t.Fatalf("Failed to unmarshal sell order response: %v", err)
	}

	_, buyerToken := createRandomUser(t)

	executePayload := types.ExecuteOrderPayload{
		OrderID: sellOrder.ID,
	}

	body, _ = json.Marshal(executePayload)
	req, _ = http.NewRequest("POST", "/order/execute", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buyerToken)
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var executeResponse map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &executeResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal execute order response: %v", err)
	}

	if message, exists := executeResponse["message"]; !exists || message != "Order executed successfully" {
		t.Errorf("Handler returned unexpected message: got %v want %v", message, "Order executed successfully")
	}

	time.Sleep(time.Millisecond * 100)

	req, _ = http.NewRequest("GET", fmt.Sprintf("/order/list/%d?type=sell", createdToken.ID), nil)
	req.Header.Set("Authorization", sellerToken)
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code when listing orders: got %v want %v", status, http.StatusOK)
	}

	var orders []*types.Order
	err = json.Unmarshal(rr.Body.Bytes(), &orders)
	if err != nil {
		t.Fatalf("Failed to unmarshal list orders response: %v", err)
	}

	if len(orders) != 0 {
		t.Fatalf("Expected 0 order, got %d", len(orders))
	}

	req, _ = http.NewRequest("GET", fmt.Sprintf("/token/balance/%d", createdToken.ID), nil)
	req.Header.Set("Authorization", buyerToken)
	rr = httptest.NewRecorder()

	tokenHandler.RegisterRoutes(router)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code when getting buyer balance: got %v want %v", status, http.StatusOK)
	}

	var balanceResponse map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &balanceResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal balance response: %v", err)
	}

	buyerBalance, ok := new(big.Int).SetString(balanceResponse["balance"], 10)
	if !ok {
		t.Fatalf("Failed to parse buyer balance")
	}

	if buyerBalance.String() != "100" {
		t.Errorf("Unexpected buyer balance: got %v want %v", buyerBalance.String(), "100")
	}
}
