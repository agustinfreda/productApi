package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"

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
	sortParam := r.URL.Query().Get("sort")

	if sortParam == "price" {
		sort.Slice(products, func(i, j int) bool {
			return products[i].Price > products[j].Price // Orden descendente
		})
	}

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

func getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["id"])

	if err != nil {
		fmt.Fprintf(w, "Invalid ID")
		return
	}

	for _, product := range products {
		if product.ID == productID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(product)
		}
	}
}

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["id"])

	if err != nil {
		fmt.Fprintf(w, "Invalid ID")
		return
	}

	for i, product := range products {
		if product.ID == productID {
			products = append(products[:i], products[i+1:]...)
			fmt.Fprintf(w, "The product with ID %v has been removed succesfully", productID)
		}
	}

}

func updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["id"])

	var updatedProduct Product

	if err != nil {
		fmt.Fprintf(w, "Invalid ID")
	}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Please Enter Valid Data")
	}

	json.Unmarshal(reqBody, &updatedProduct)

	for i, product := range products {
		if product.ID == productID {
			products = append(products[:i], products[i+1:]...)
			updatedProduct.ID = productID
			products = append(products, updatedProduct)

			fmt.Fprintf(w, "The product with ID %v has been updated succesfully", productID)
		}
	}
}

func sellProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	sellParam := r.URL.Query().Get("amount")
	amount := 1 // Por defecto se vende 1

	if sellParam != "" {
		amount, err = strconv.Atoi(sellParam)
		if err != nil || amount <= 0 {
			http.Error(w, "Invalid amount", http.StatusBadRequest)
			return
		}
	}

	// Buscar producto por índice y modificar el stock
	found := false
	for i := range products {
		if products[i].ID == productID {
			if products[i].Stock >= amount {
				products[i].Stock -= amount
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "You sold %d units of product with ID %v", amount, productID)
			} else {
				http.Error(w, "Not enough stock", http.StatusConflict)
			}
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Product not found", http.StatusNotFound)
	}
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
	router.HandleFunc("/products/{id}", getProduct).Methods("GET")
	router.HandleFunc("/products/{id}", deleteProduct).Methods("DELETE")
	router.HandleFunc("/products/{id}", updateProduct).Methods("PUT")
	router.HandleFunc("/products/{id}/sell", sellProduct).Methods("PUT")

	// Server HTTP
	log.Fatal(http.ListenAndServe(":4567", router))
}
