package main

import (
	"fmt"
	"go-payment-platform/internal/httpapi"
	"go-payment-platform/internal/payment"
	"go-payment-platform/internal/storage/memory"
	"net/http"
)

func main() {
	store := memory.NewStore()
	p := payment.NewService(store)
	router := httpapi.NewRouter(p)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println("Server error", err)
	}
}
