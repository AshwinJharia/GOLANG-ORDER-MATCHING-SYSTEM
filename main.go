package main

import (
	"log"
	"net/http"
	"order-matching-engine/database"
	"order-matching-engine/handlers"
	"order-matching-engine/services"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize matching engine
	engine := services.NewMatchingEngine()

	// Initialize handlers
	orderHandler := handlers.NewOrderHandler(engine)
	tradeHandler := handlers.NewTradeHandler()

	// Setup routes
	router := mux.NewRouter()
	
	// Order endpoints with method validation
	router.HandleFunc("/orders", orderHandler.PlaceOrder).Methods("POST")
	router.HandleFunc("/orders", methodNotAllowed).Methods("GET", "PUT", "DELETE", "PATCH")
	router.HandleFunc("/orders/{id}", orderHandler.GetOrder).Methods("GET")
	router.HandleFunc("/orders/{id}", orderHandler.CancelOrder).Methods("DELETE")
	router.HandleFunc("/orders/{id}", methodNotAllowed).Methods("POST", "PUT", "PATCH")
	router.HandleFunc("/orderbook", orderHandler.GetOrderBook).Methods("GET")
	router.HandleFunc("/orderbook", methodNotAllowed).Methods("POST", "PUT", "DELETE", "PATCH")
	
	// Trade endpoints with method validation
	router.HandleFunc("/trades", tradeHandler.GetTrades).Methods("GET")
	router.HandleFunc("/trades", methodNotAllowed).Methods("POST", "PUT", "DELETE", "PATCH")

	// Health check with method validation
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	router.HandleFunc("/health", methodNotAllowed).Methods("POST", "PUT", "DELETE", "PATCH")

	log.Println("Order Matching Engine starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"success":false,"error":"Method not allowed"}`))
}