package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
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

type Credentials struct {
	Username string `json:Username`
	Password string `json:Password`
}

var secretKey = []byte("miClaveSecreta")

type allCredentials []Credentials

var usersRegistered = allCredentials{
	{
		Username: "agustinfreda",
		Password: "12345678",
	},
	{
		Username: "ri-ma1",
		Password: "josefue",
	},
}

// PRODUCT FUNCTIONS
func getProducts(w http.ResponseWriter, r *http.Request) {
	sortParam := r.URL.Query().Get("sort")
	searchParam := r.URL.Query().Get("name")

	var searchSlice []Product
	for _, product := range products {
		// Filtrar productos que contienen el nombre buscado (insensible a mayúsculas/minúsculas)
		if strings.Contains(strings.ToLower(product.Name), strings.ToLower(searchParam)) {
			searchSlice = append(searchSlice, product)
		}
	}

	// Si se especifica el parámetro "sort", ordenar los resultados filtrados
	if sortParam == "price" {
		sort.Slice(searchSlice, func(i, j int) bool {
			return searchSlice[i].Price > searchSlice[j].Price // Orden descendente
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searchSlice) // Devolver los productos filtrados y ordenados
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

func generateJWT(username string) (string, error) {
	// Definir los claims del token
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(), // El token expira en 72 horas
	}

	// Crear el token con los claims y la firma
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmar el token usando la clave secreta
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	for _, user := range usersRegistered {
		if user.Username == creds.Username && user.Password == creds.Password {
			token, err := generateJWT(creds.Username)
			if err != nil {
				http.Error(w, "Could not generate token", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"token": token})
			return
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized", "message": "Missing or invalid token"})

}

var revokedTokens = make(map[string]bool) // Aquí guardamos los tokens revocados

func logout(w http.ResponseWriter, r *http.Request) {
	// Obtener el token del header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}

	// El token es la segunda parte (después de "Bearer")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		http.Error(w, "Authorization header format must be Bearer <token>", http.StatusUnauthorized)
		return
	}
	tokenString := parts[1]

	// Agregar el token a la lista de revocados
	revokedTokens[tokenString] = true

	// Responder con un mensaje de éxito
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout successful"})
}

// Función para verificar si el token está revocado
func isTokenRevoked(tokenString string) bool {
	return revokedTokens[tokenString]
}

// Middleware que verifica si el token está revocado
func tokenVerifyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			http.Error(w, "Authorization header format must be Bearer <token>", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		if isTokenRevoked(tokenString) {
			http.Error(w, "Token has been revoked", http.StatusUnauthorized)
			return
		}

		// Continuar con la validación del token
		_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrNotSupported
			}
			return secretKey, nil
		})

		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	})
}

func main() {
	// Enrutador
	router := mux.NewRouter().StrictSlash(true)

	// Rutas públicas
	router.HandleFunc("/", indexRoute)
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/logout", logout).Methods("POST")

	// Rutas protegidas (requieren autenticación)
	router.HandleFunc("/products", tokenVerifyMiddleware(getProducts)).Methods("GET")
	router.HandleFunc("/products", tokenVerifyMiddleware(createProduct)).Methods("POST")
	router.HandleFunc("/products/{id}", tokenVerifyMiddleware(getProduct)).Methods("GET")
	router.HandleFunc("/products/{id}", tokenVerifyMiddleware(deleteProduct)).Methods("DELETE")
	router.HandleFunc("/products/{id}", tokenVerifyMiddleware(updateProduct)).Methods("PUT")
	router.HandleFunc("/products/{id}/sell", tokenVerifyMiddleware(sellProduct)).Methods("PUT")

	// Server HTTP
	log.Fatal(http.ListenAndServe(":4567", router))
}
