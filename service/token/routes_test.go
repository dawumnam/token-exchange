package token

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

	"github.com/dawumnam/token-trader/config"
	"github.com/dawumnam/token-trader/db"
	"github.com/dawumnam/token-trader/service/user"
	"github.com/dawumnam/token-trader/types"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var testDB *sql.DB
var tokenHandler *Handler
var userHandler *user.Handler

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

	tokenRepo := NewTokenRepository(testDB)
	userRepo := user.NewRepository(testDB)
	txManager := db.NewTxManager(testDB)

	tokenHandler = NewHandler(tokenRepo, userRepo, txManager)
	userHandler = user.NewHandler(userRepo)

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

func TestHandleIssueToken(t *testing.T) {
	_, token := createRandomUser(t)

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
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response types.Token
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Name != payload.Name || response.Symbol != payload.Symbol {
		t.Errorf("Handler returned unexpected body: %+v", response)
	}
}

func TestHandleGetBalance(t *testing.T) {
	_, token := createRandomUser(t)

	issuePayload := types.IssueTokenPayload{
		Name:          "Balance Test Token",
		Symbol:        "BTT",
		InitialSupply: "1000",
	}

	body, _ := json.Marshal(issuePayload)
	req, _ := http.NewRequest("POST", "/token/issue", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	tokenHandler.RegisterRoutes(router)
	router.ServeHTTP(rr, req)

	var issuedToken types.Token
	err := json.Unmarshal(rr.Body.Bytes(), &issuedToken)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	req, _ = http.NewRequest("GET", fmt.Sprintf("/token/balance/%d", issuedToken.ID), nil)
	req.Header.Set("Authorization", token)
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var balanceResponse map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &balanceResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	balance, ok := new(big.Int).SetString(balanceResponse["balance"], 10)
	if !ok {
		t.Fatalf("Failed to parse balance")
	}

	expectedBalance, _ := new(big.Int).SetString(issuePayload.InitialSupply, 10)
	if balance.Cmp(expectedBalance) != 0 {
		t.Errorf("Unexpected balance: got %v want %v", balance, expectedBalance)
	}
}

func TestHandleListTokens(t *testing.T) {
	_, token := createRandomUser(t)

	for i := 0; i < 2; i++ {
		issuePayload := types.IssueTokenPayload{
			Name:          fmt.Sprintf("List Test Token %d", i),
			Symbol:        fmt.Sprintf("LTT%d", i),
			InitialSupply: "1000",
		}

		body, _ := json.Marshal(issuePayload)
		req, _ := http.NewRequest("POST", "/token/issue", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		tokenHandler.RegisterRoutes(router)
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
		}
	}

	req, _ := http.NewRequest("GET", "/token/list", nil)
	req.Header.Set("Authorization", token)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	tokenHandler.RegisterRoutes(router)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var tokens []*types.Token
	err := json.Unmarshal(rr.Body.Bytes(), &tokens)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(tokens) != 2 {
		t.Errorf("Unexpected number of tokens: got %v want %v", len(tokens), 2)
	}
}
