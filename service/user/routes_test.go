package user

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dawumnam/token-trader/config"
	"github.com/dawumnam/token-trader/db"
	"github.com/dawumnam/token-trader/types"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var testDB *sql.DB
var handler *Handler

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

	userRepo := NewRepository(testDB)
	handler = NewHandler(userRepo)

	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

func createRandomUser() types.RegisterUserPayload {
	return types.RegisterUserPayload{
		Email:     fmt.Sprintf("user_%s@example.com", uuid.New().String()),
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}
}

func TestHandleRegister(t *testing.T) {
	payload := createRandomUser()

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/register", handler.HandleRegister)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}
}

func TestHandleLogin(t *testing.T) {
	registerPayload := createRandomUser()

	body, _ := json.Marshal(registerPayload)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/register", handler.HandleRegister)
	router.ServeHTTP(rr, req)

	loginPayload := types.LoginUserPayload{
		Email:    registerPayload.Email,
		Password: registerPayload.Password,
	}

	body, _ = json.Marshal(loginPayload)
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	router = mux.NewRouter()
	router.HandleFunc("/login", handler.handleLogin)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if _, exists := response["token"]; !exists {
		t.Errorf("Response does not contain token")
	}
}

func TestHandleLogout(t *testing.T) {
	registerPayload := createRandomUser()

	body, _ := json.Marshal(registerPayload)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/register", handler.HandleRegister)
	router.ServeHTTP(rr, req)

	var registerResponse map[string]string
	json.Unmarshal(rr.Body.Bytes(), &registerResponse)
	token := registerResponse["token"]

	req, _ = http.NewRequest("POST", "/logout", nil)
	req.Header.Set("Authorization", token)
	rr = httptest.NewRecorder()

	router = mux.NewRouter()
	router.HandleFunc("/logout", handler.handleLogout)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var logoutResponse map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &logoutResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if message, exists := logoutResponse["message"]; !exists || message != "Successfully logged out" {
		t.Errorf("Handler returned unexpected message: got %v want %v", message, "Successfully logged out")
	}
}
