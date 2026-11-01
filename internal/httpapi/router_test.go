package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"go-payment-platform/internal/config"
	"go-payment-platform/internal/payment"
	"go-payment-platform/internal/storage/postgres"

	"github.com/joho/godotenv"
)

func setupTestServer(t *testing.T) http.Handler {
	t.Helper()

	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("load .env: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	testDBName := strings.TrimSpace(os.Getenv("TEST_DB_NAME"))
	if testDBName != "payment_platform_test" {
		t.Fatalf("unsafe test database name: %q", testDBName)
	}

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		testDBName,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store, err := postgres.NewStore(ctx, connString)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	t.Cleanup(store.Close)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := payment.NewService(store)

	return NewRouter(service, logger)
}

func TestCreateAccount(t *testing.T) {
	server := setupTestServer(t)

	body := bytes.NewBufferString(`{"owner":"Test User"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d; body: %s",
			http.StatusCreated, rec.Code, rec.Body.String())
	}

	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response["owner"] != "Test User" {
		t.Fatalf("expected owner %q, got %v", "Test User", response["owner"])
	}
}

func TestGetAccount(t *testing.T) {
	server := setupTestServer(t)

	// Сначала создаём аккаунт.
	createBody := bytes.NewBufferString(`{"owner":"Get Test User"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/accounts", createBody)
	createRec := httptest.NewRecorder()
	server.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create account: expected status %d, got %d; body: %s",
			http.StatusCreated, createRec.Code, createRec.Body.String())
	}

	var account map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &account); err != nil {
		t.Fatalf("decode created account: %v", err)
	}

	id, ok := account["id"].(float64)
	if !ok || id <= 0 {
		t.Fatalf("invalid account id: %v", account["id"])
	}

	// Затем получаем его по ID.
	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/accounts/%.0f", id),
		nil,
	)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body: %s",
			http.StatusOK, rec.Code, rec.Body.String())
	}
}

func TestDeposit(t *testing.T) {
	server := setupTestServer(t)

	createBody := bytes.NewBufferString(`{"owner":"Deposit Test User"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/accounts", createBody)
	createRec := httptest.NewRecorder()
	server.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create account: %s", createRec.Body.String())
	}

	var account map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &account); err != nil {
		t.Fatal(err)
	}

	id := int(account["id"].(float64))

	depositBody := bytes.NewBufferString(fmt.Sprintf(
		`{"account_id":%d,"amount":10000}`,
		id,
	))
	req := httptest.NewRequest(http.MethodPost, "/deposits", depositBody)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body: %s",
			http.StatusOK, rec.Code, rec.Body.String())
	}
}

func TestTransfer(t *testing.T) {
	server := setupTestServer(t)

	createAccount := func(owner string) int {
		body := bytes.NewBufferString(fmt.Sprintf(`{"owner":%q}`, owner))
		req := httptest.NewRequest(http.MethodPost, "/accounts", body)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("create account: %s", rec.Body.String())
		}

		var account map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &account); err != nil {
			t.Fatal(err)
		}

		return int(account["id"].(float64))
	}

	fromID := createAccount("Transfer Sender")
	toID := createAccount("Transfer Receiver")

	// Пополняем счёт отправителя.
	depositBody := bytes.NewBufferString(fmt.Sprintf(
		`{"account_id":%d,"amount":20000}`,
		fromID,
	))
	depositReq := httptest.NewRequest(http.MethodPost, "/deposits", depositBody)
	depositRec := httptest.NewRecorder()
	server.ServeHTTP(depositRec, depositReq)

	if depositRec.Code != http.StatusOK {
		t.Fatalf("deposit: %s", depositRec.Body.String())
	}

	// Переводим часть средств получателю.
	transferBody := bytes.NewBufferString(fmt.Sprintf(
		`{"from_id":%d,"to_id":%d,"amount":5000}`,
		fromID,
		toID,
	))
	req := httptest.NewRequest(http.MethodPost, "/transfers", transferBody)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d; body: %s",
			http.StatusNoContent, rec.Code, rec.Body.String())
	}
}

func TestGetTransactions(t *testing.T) {
	server := setupTestServer(t)

	createBody := bytes.NewBufferString(`{"owner":"Transactions Test User"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/accounts", createBody)
	createRec := httptest.NewRecorder()
	server.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create account: %s", createRec.Body.String())
	}

	var account map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &account); err != nil {
		t.Fatal(err)
	}

	id := int(account["id"].(float64))

	depositBody := bytes.NewBufferString(fmt.Sprintf(
		`{"account_id":%d,"amount":10000}`,
		id,
	))
	depositReq := httptest.NewRequest(http.MethodPost, "/deposits", depositBody)
	depositRec := httptest.NewRecorder()
	server.ServeHTTP(depositRec, depositReq)

	if depositRec.Code != http.StatusOK {
		t.Fatalf("deposit: %s", depositRec.Body.String())
	}

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/accounts/%d/transactions?offset=0&limit=10", id),
		nil,
	)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body: %s",
			http.StatusOK, rec.Code, rec.Body.String())
	}
}
