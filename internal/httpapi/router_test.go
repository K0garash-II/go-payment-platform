package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-payment-platform/internal/payment"
	"go-payment-platform/internal/storage/memory"
	"go-payment-platform/internal/transaction"
)

func setupTestServer() http.Handler {
	store := memory.NewStore()
	service := payment.NewService(store)
	return NewRouter(store, service)
}

func sendRequest(
	t *testing.T,
	router http.Handler,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(
		method,
		path,
		strings.NewReader(body),
	)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	return rec
}

func TestPaymentAPI(t *testing.T) {
	router := setupTestServer()

	// 1. Создаём счёт.
	res := sendRequest(
		t, router, http.MethodPost,
		"/accounts", `{"owner":"Alex"}`,
	)

	if res.Code != http.StatusCreated {
		t.Fatalf("create account: got %d, want %d; body: %s",
			res.Code, http.StatusCreated, res.Body.String())
	}

	// 2. Первое пополнение: 1500 центов.
	res = sendRequest(
		t, router, http.MethodPost,
		"/accounts/1/deposit", `{"amount":1500}`,
	)

	if res.Code != http.StatusOK {
		t.Fatalf("first deposit: got %d, want %d",
			res.Code, http.StatusOK)
	}

	// 3. Второе пополнение: 500 центов.
	res = sendRequest(
		t, router, http.MethodPost,
		"/accounts/1/deposit", `{"amount":500}`,
	)

	if res.Code != http.StatusOK {
		t.Fatalf("second deposit: got %d, want %d",
			res.Code, http.StatusOK)
	}

	// 4. Проверяем итоговый баланс: 2000 центов.
	res = sendRequest(
		t, router, http.MethodGet,
		"/accounts/1", "",
	)

	if res.Code != http.StatusOK {
		t.Fatalf("get account: got %d, want %d",
			res.Code, http.StatusOK)
	}

	var acc struct {
		Balance int64 `json:"balance"`
	}

	if err := json.NewDecoder(res.Body).Decode(&acc); err != nil {
		t.Fatalf("decode account: %v", err)
	}

	if acc.Balance != 2000 {
		t.Fatalf("balance: got %d, want 2000", acc.Balance)
	}

	// 5. Проверяем историю транзакций.
	res = sendRequest(
		t, router, http.MethodGet,
		"/accounts/1/transactions", "",
	)

	if res.Code != http.StatusOK {
		t.Fatalf("get history: got %d, want %d; body: %s",
			res.Code, http.StatusOK, res.Body.String())
	}

	var transactions []transaction.Transaction

	if err := json.NewDecoder(res.Body).Decode(&transactions); err != nil {
		t.Fatalf("decode transactions: %v", err)
	}

	if len(transactions) != 2 {
		t.Fatalf("transaction count: got %d, want 2", len(transactions))
	}

	if transactions[0].Amount != 1500 {
		t.Errorf("first transaction amount: got %d, want 1500",
			transactions[0].Amount)
	}

	if transactions[1].Amount != 500 {
		t.Errorf("second transaction amount: got %d, want 500",
			transactions[1].Amount)
	}

	// 6. Проверяем пополнение несуществующего счёта.
	res = sendRequest(
		t, router, http.MethodPost,
		"/accounts/999/deposit", `{"amount":500}`,
	)

	if res.Code != http.StatusNotFound {
		t.Errorf("missing account: got %d, want %d",
			res.Code, http.StatusNotFound)
	}

	// 7. Отрицательное пополнение должно быть отклонено.
	res = sendRequest(
		t, router, http.MethodPost,
		"/accounts/1/deposit", `{"amount":-100}`,
	)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("negative deposit: got %d, want %d",
			res.Code, http.StatusBadRequest)
	}

	// 8. После ошибки баланс не должен измениться.
	res = sendRequest(
		t, router, http.MethodGet,
		"/accounts/1", "",
	)

	if err := json.NewDecoder(res.Body).Decode(&acc); err != nil {
		t.Fatalf("decode account after invalid deposit: %v", err)
	}

	if acc.Balance != 2000 {
		t.Errorf("balance after invalid deposit: got %d, want 2000",
			acc.Balance)
	}

	// 9. Проверяем, что ошибочная операция не попала в историю.
	res = sendRequest(
		t, router, http.MethodGet,
		"/accounts/1/transactions", "",
	)

	transactions = nil

	if err := json.NewDecoder(res.Body).Decode(&transactions); err != nil {
		t.Fatalf("decode history after invalid deposit: %v", err)
	}

	if len(transactions) != 2 {
		t.Errorf("history after invalid deposit: got %d transactions, want 2",
			len(transactions))
	}
}
