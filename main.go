package main

import (
	"encoding/json"
	"fmt"
	"io"
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

func createProduct(w http.ResponseWriter, r *http.Request) {
	var newProduct Product
	//guardar datos de la request
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Insert a Valid Product")
	}

	//asignando la info recibida a la variable.
	json.Unmarshal(reqBody, &newProduct)

	newProduct.ID = len(products) + 1
	products = append(products, newProduct)

	//respuesta al cliente
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newProduct)
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
	router.HandleFunc("/products", createProduct).Methods("POST")

	// Server HTTP
	log.Fatal(http.ListenAndServe(":4567", router))
}
