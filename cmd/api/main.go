package main

import (
	"fmt"
	"net/http"
	"go-payment-platform/internal/httpapi"
	"go-payment-platform/internal/storage/memory"
)



func main() {
	store := memory.NewStore()
	router := httpapi.NewRouter(store)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println("Server error", err)
	}
}
