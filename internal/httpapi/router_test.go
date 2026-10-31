package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"go-payment-platform/internal/payment"
	"go-payment-platform/internal/storage/memory"
	"go-payment-platform/internal/transaction"
)

func setupTestServer() http.Handler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	store := memory.NewStore()
	service := payment.NewService(store)
	return NewRouter(service, logger)
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

	// 1. Создаём первый счёт.
	res := sendRequest(
		t, router, http.MethodPost,
		"/accounts", `{"owner":"Alex"}`,
	)

	if res.Code != http.StatusCreated {
		t.Fatalf("create first account: got %d, want %d; body: %s",
			res.Code, http.StatusCreated, res.Body.String())
	}

	// 2. Создаём второй счёт для проверки переводов.
	res = sendRequest(
		t, router, http.MethodPost,
		"/accounts", `{"owner":"Bob"}`,
	)

	if res.Code != http.StatusCreated {
		t.Fatalf("create second account: got %d, want %d; body: %s",
			res.Code, http.StatusCreated, res.Body.String())
	}

	// 3. Первое пополнение счёта Alex: 1500 центов.
	res = sendRequest(
		t, router, http.MethodPost,
		"/deposits", `{"account_id":1,"amount":1500}`,
	)

	if res.Code != http.StatusOK {
		t.Fatalf("first deposit: got %d, want %d; body: %s",
			res.Code, http.StatusOK, res.Body.String())
	}

	// 4. Второе пополнение счёта Alex: 500 центов.
	res = sendRequest(
		t, router, http.MethodPost,
		"/deposits", `{"account_id":1,"amount":500}`,
	)

	if res.Code != http.StatusOK {
		t.Fatalf("second deposit: got %d, want %d; body: %s",
			res.Code, http.StatusOK, res.Body.String())
	}

	// 5. Проверяем баланс Alex: 2000 центов.
	res = sendRequest(
		t, router, http.MethodGet,
		"/accounts/1", "",
	)

	if res.Code != http.StatusOK {
		t.Fatalf("get Alex account: got %d, want %d",
			res.Code, http.StatusOK)
	}

	var alex struct {
		Balance int64 `json:"balance"`
	}

	if err := json.NewDecoder(res.Body).Decode(&alex); err != nil {
		t.Fatalf("decode Alex account: %v", err)
	}

	if alex.Balance != 2000 {
		t.Fatalf("Alex balance: got %d, want 2000", alex.Balance)
	}

	// 6. Переводим 700 центов от Alex к Bob.
	res = sendRequest(
		t, router, http.MethodPost,
		"/transfers", `{"from_id":1,"to_id":2,"amount":700}`,
	)

	if res.Code != http.StatusNoContent {
		t.Fatalf("successful transfer: got %d, want %d; body: %s",
			res.Code, http.StatusOK, res.Body.String())
	}

	// 7. Проверяем баланс Alex после перевода.
	res = sendRequest(
		t, router, http.MethodGet,
		"/accounts/1", "",
	)

	if res.Code != http.StatusOK {
		t.Fatalf("get Alex balance after transfer: got %d, want %d",
			res.Code, http.StatusOK)
	}

	if err := json.NewDecoder(res.Body).Decode(&alex); err != nil {
		t.Fatalf("decode Alex balance after transfer: %v", err)
	}

	if alex.Balance != 1300 {
		t.Fatalf("Alex balance after transfer: got %d, want 1300",
			alex.Balance)
	}

	// 8. Проверяем баланс Bob после перевода.
	res = sendRequest(
		t, router, http.MethodGet,
		"/accounts/2", "",
	)

	if res.Code != http.StatusOK {
		t.Fatalf("get Bob account: got %d, want %d",
			res.Code, http.StatusOK)
	}

	var bob struct {
		Balance int64 `json:"balance"`
	}

	if err := json.NewDecoder(res.Body).Decode(&bob); err != nil {
		t.Fatalf("decode Bob account: %v", err)
	}

	if bob.Balance != 700 {
		t.Fatalf("Bob balance after transfer: got %d, want 700",
			bob.Balance)
	}

	// 9. Проверяем историю пополнений Alex.
	res = sendRequest(
		t, router, http.MethodGet,
		"/accounts/1/transactions", "",
	)

	if res.Code != http.StatusOK {
		t.Fatalf("get Alex history: got %d, want %d; body: %s",
			res.Code, http.StatusOK, res.Body.String())
	}

	var transactions []transaction.Transaction

	if err := json.NewDecoder(res.Body).Decode(&transactions); err != nil {
		t.Fatalf("decode Alex transactions: %v", err)
	}

	// Два пополнения и один перевод.
	if len(transactions) != 3 {
		t.Fatalf("Alex transaction count: got %d, want 3",
			len(transactions))
	}

	if transactions[0].Amount != 1500 {
		t.Errorf("first deposit amount: got %d, want 1500",
			transactions[0].Amount)
	}

	if transactions[1].Amount != 500 {
		t.Errorf("second deposit amount: got %d, want 500",
			transactions[1].Amount)
	}

	// 10. Перевод при недостаточном балансе должен быть отклонён.
	res = sendRequest(
		t, router, http.MethodPost,
		"/transfers", `{"from_id":1,"to_id":2,"amount":5000}`,
	)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("insufficient funds: got %d, want %d",
			res.Code, http.StatusBadRequest)
	}

	// 11. Перевод на тот же счёт должен быть отклонён.
	res = sendRequest(
		t, router, http.MethodPost,
		"/transfers", `{"from_id":1,"to_id":1,"amount":100}`,
	)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("same account transfer: got %d, want %d",
			res.Code, http.StatusBadRequest)
	}

	// 12. Перевод с несуществующего счёта должен вернуть 404.
	res = sendRequest(
		t, router, http.MethodPost,
		"/transfers", `{"from_id":999,"to_id":2,"amount":100}`,
	)

	if res.Code != http.StatusNotFound {
		t.Fatalf("transfer from missing account: got %d, want %d",
			res.Code, http.StatusNotFound)
	}

	// 13. Пополнение несуществующего счёта должно вернуть 404.
	res = sendRequest(
		t, router, http.MethodPost,
		"/deposits", `{"account_id":999,"amount":500}`,
	)

	if res.Code != http.StatusNotFound {
		t.Fatalf("deposit to missing account: got %d, want %d",
			res.Code, http.StatusNotFound)
	}

	// 14. Отрицательное пополнение должно быть отклонено.
	res = sendRequest(
		t, router, http.MethodPost,
		"/deposits", `{"account_id":1,"amount":-100}`,
	)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("negative deposit: got %d, want %d",
			res.Code, http.StatusBadRequest)
	}

	// 15. После неудачных операций балансы не должны измениться.
	res = sendRequest(
		t, router, http.MethodGet,
		"/accounts/1", "",
	)

	if err := json.NewDecoder(res.Body).Decode(&alex); err != nil {
		t.Fatalf("decode Alex after failed operations: %v", err)
	}

	if alex.Balance != 1300 {
		t.Errorf("Alex balance after failed operations: got %d, want 1300",
			alex.Balance)
	}

	res = sendRequest(
		t, router, http.MethodGet,
		"/accounts/2", "",
	)

	if err := json.NewDecoder(res.Body).Decode(&bob); err != nil {
		t.Fatalf("decode Bob after failed operations: %v", err)
	}

	if bob.Balance != 700 {
		t.Errorf("Bob balance after failed operations: got %d, want 700",
			bob.Balance)
	}

	// 16. Неудачные операции не должны добавлять транзакции Alex.
	res = sendRequest(
		t, router, http.MethodGet,
		"/accounts/1/transactions", "",
	)

	transactions = nil

	if err := json.NewDecoder(res.Body).Decode(&transactions); err != nil {
		t.Fatalf("decode Alex history after failed operations: %v", err)
	}

	if len(transactions) != 3 {
		t.Errorf("Alex history after failed operations: got %d transactions, want 3",
			len(transactions))
	}
}
