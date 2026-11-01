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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type testEnv struct {
	server     http.Handler
	store      *postgres.Store
	cleanupDB  *pgxpool.Pool
	accountIDs []int64
}

func setupTestServer(t *testing.T) *testEnv {
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

	cleanupDB, err := pgxpool.New(ctx, connString)
	if err != nil {
		store.Close()
		t.Fatalf("connect cleanup database: %v", err)
	}

	if err := cleanupDB.Ping(ctx); err != nil {
		cleanupDB.Close()
		store.Close()
		t.Fatalf("ping cleanup database: %v", err)
	}

	env := &testEnv{
		store:     store,
		cleanupDB: cleanupDB,
	}

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if len(env.accountIDs) > 0 {
			_, err := env.cleanupDB.Exec(
				cleanupCtx,
				`DELETE FROM transactions
             WHERE account_id = ANY($1::bigint[])`,
				env.accountIDs,
			)
			if err != nil {
				t.Errorf("cleanup transactions: %v", err)
			}

			_, err = env.cleanupDB.Exec(
				cleanupCtx,
				`DELETE FROM accounts
             WHERE id = ANY($1::bigint[])`,
				env.accountIDs,
			)
			if err != nil {
				t.Errorf("cleanup accounts: %v", err)
			}
		}

		env.cleanupDB.Close()
		env.store.Close()
	})

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := payment.NewService(store)

	env.server = NewRouter(service, logger)

	return env
}

func (env *testEnv) createAccount(t *testing.T, owner string) int {
	t.Helper()

	body := bytes.NewBufferString(
		fmt.Sprintf(`{"owner":%q}`, owner),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/accounts",
		body,
	)

	rec := httptest.NewRecorder()
	env.server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"create account: expected status %d, got %d; body: %s",
			http.StatusCreated,
			rec.Code,
			rec.Body.String(),
		)
	}

	var response struct {
		ID int `json:"id"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode created account: %v", err)
	}

	if response.ID <= 0 {
		t.Fatalf("invalid account ID: %d", response.ID)
	}

	env.accountIDs = append(env.accountIDs, int64(response.ID))

	return response.ID
}

func TestCreateAccount(t *testing.T) {
	env := setupTestServer(t)

	env.createAccount(t, "Test User")
}

func TestGetAccount(t *testing.T) {
	env := setupTestServer(t)

	id := env.createAccount(t, "Get Test User")

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/accounts/%d", id),
		nil,
	)

	rec := httptest.NewRecorder()
	env.server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestDeposit(t *testing.T) {
	env := setupTestServer(t)

	id := env.createAccount(t, "Deposit Test User")

	body := bytes.NewBufferString(fmt.Sprintf(
		`{"account_id":%d,"amount":10000}`,
		id,
	))

	req := httptest.NewRequest(
		http.MethodPost,
		"/deposits",
		body,
	)

	rec := httptest.NewRecorder()
	env.server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestTransfer(t *testing.T) {
	env := setupTestServer(t)

	fromID := env.createAccount(t, "Transfer Sender")
	toID := env.createAccount(t, "Transfer Receiver")

	depositBody := bytes.NewBufferString(fmt.Sprintf(
		`{"account_id":%d,"amount":20000}`,
		fromID,
	))

	depositReq := httptest.NewRequest(
		http.MethodPost,
		"/deposits",
		depositBody,
	)

	depositRec := httptest.NewRecorder()
	env.server.ServeHTTP(depositRec, depositReq)

	if depositRec.Code != http.StatusOK {
		t.Fatalf("deposit: %s", depositRec.Body.String())
	}

	transferBody := bytes.NewBufferString(fmt.Sprintf(
		`{"from_id":%d,"to_id":%d,"amount":5000}`,
		fromID,
		toID,
	))

	req := httptest.NewRequest(
		http.MethodPost,
		"/transfers",
		transferBody,
	)

	rec := httptest.NewRecorder()
	env.server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusNoContent,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestGetTransactions(t *testing.T) {
	env := setupTestServer(t)

	id := env.createAccount(t, "Transactions Test User")

	depositBody := bytes.NewBufferString(fmt.Sprintf(
		`{"account_id":%d,"amount":10000}`,
		id,
	))

	depositReq := httptest.NewRequest(
		http.MethodPost,
		"/deposits",
		depositBody,
	)

	depositRec := httptest.NewRecorder()
	env.server.ServeHTTP(depositRec, depositReq)

	if depositRec.Code != http.StatusOK {
		t.Fatalf("deposit: %s", depositRec.Body.String())
	}

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/accounts/%d/transactions?offset=0&limit=10",
			id,
		),
		nil,
	)

	rec := httptest.NewRecorder()
	env.server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}
