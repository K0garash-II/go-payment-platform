package main

import (
	"fmt"
	"net/http"
	"github.com/K0garash-II/go-payment-platform/internal/httpapi"
)

func main() {
	router := httpapi.NewRouter()

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println("Server error", err)
	}
}
