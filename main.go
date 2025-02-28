package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Product struct {
	ID    int    `json:ID`
	Name  string `json:Name`
	Price int    `json:Price`
	Stock int    `json:Stock`
}

type allProducts []Product

var products = allProducts{
	{
		ID:    1,
		Name:  "Product One",
		Price: 10,
		Stock: 250,
	},
	{
		ID:    2,
		Name:  "Product Two",
		Price: 100,
		Stock: 20,
	},
}

// PRODUCT FUNCTIONS
func getProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func indexRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to my API")
}

func main() {
	// Enrutador
	router := mux.NewRouter().StrictSlash(true)

	// Rutas
	router.HandleFunc("/", indexRoute)
	router.HandleFunc("/products", getProducts).Methods("GET")

	// Server HTTP
	log.Fatal(http.ListenAndServe(":4567", router))
}
