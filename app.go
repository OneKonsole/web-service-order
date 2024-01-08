package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"

	"encoding/json"
	"net/http"
	"strconv"

	oko "github.com/OneKonsole/order-model"

	"github.com/gorilla/mux"

	_ "github.com/lib/pq"

	"github.com/go-playground/validator/v10"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// ===========================================================================================================
// Initialize database and http server for the order service
// Used on:
//
//	a (*App) : App struct containing the service necessary items
//
// Parameters:
//
//	user (string) : Database user
//	password (string) : Database password
//	dbName (string) : Database name
//
// Examples:
//
//	a.Initialize("testuser","testpassword","mydb")
//
// ===========================================================================================================
func (a *App) Initialize(user string, password string, dbname string) {
	connectionString := fmt.Sprintf("postgresql://%s:%s@localhost/%s?sslmode=disable", user, password, dbname)
	// connectionString := fmt.Sprintf("postgresql://%s:%s@my-postgresql.provisioning.svc.cluster.local/%s?sslmode=disable", user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

// ===========================================================================================================
// Runs the HTTP server
//
// Used on:
//
//	a (*App) : App struct containing the service necessary items
//
// Parameters:
//
//	addr (string): Full URL to use for the server
//
// Examples:
//
//	a.Run("localhost:8010")
//
// ===========================================================================================================
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(":8010", a.Router))
}

// ===========================================================================================================
// Used as a backend for GET HTTP route /order/x to retrieve information about an order
//
// Used on:
//
//	a (*App) : App struct containing the service necessary items
//
// Parameters:
//
//	w (http.ResponseWriter) : Helper to create HTTP responses
//	r (*http.Request) : HTTP request used to launch this function
//
// Examples:
//
//	a.getOrder(w, &r)
//
// ===========================================================================================================
func (a *App) getOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	o := oko.Order{ID: id}
	if err := o.GetOrder(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Order not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, o)
}

// ===========================================================================================================
// Function called by GET HTTP route /orders that retrieves every created order
//
// Used on:
//
//	a (*App) : App struct containing the service necessary items
//
// Parameters:
//
//	w (http.ResponseWriter) : Helper to create HTTP responses
//	r (*http.Request) : HTTP request used to launch this function
//
// Examples:
//
//	a.getOrders(w, &r)
//
// ===========================================================================================================
func (a *App) getOrders(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	orders, err := oko.GetOrders(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, orders)
}

// ===========================================================================================================
// Function called by POST HTTP route /order that aims at creating a new order
// and calling provisioning producer
//
// Used on:
//
//	a (*App) : App struct containing the service necessary items
//
// Parameters:
//
//	w (http.ResponseWriter) : Helper to create HTTP responses
//	r (*http.Request) : HTTP request used to launch this function
//
// Examples:
//
//	a.createOrder(w, &r)
//
// ===========================================================================================================
func (a *App) createOrder(w http.ResponseWriter, r *http.Request) {
	var o oko.Order
	decoder := json.NewDecoder(r.Body)
	fmt.Print(r.Body)
	if err := decoder.Decode(&o); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Fields validator instance
	v := validator.New()

	// Custom validations (helpers)
	v.RegisterValidation("startswithalphanum", startsWithAlphanum)
	v.RegisterValidation("endswithalphanum", endWithAlphanum)
	v.RegisterValidation("uuid", isUUID)

	if err := v.Struct(o); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := o.CreateOrder(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	targetURL := "http://localhost:8020/produce/order"
	// targetURL := "http://sys-service-order.provisioning.svc.cluster.local:8020/produce/order"

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(o)

	resp, err := http.Post(targetURL, "application/json", buf)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()

	respondWithJSON(w, http.StatusCreated, o)
}

// ===========================================================================================================
// Function called by PUT HTTP route /order/x that aims at editing an order
//
// Used on:
//
//	a (*App) : App struct containing the service necessary items
//
// Parameters:
//
//	w (http.ResponseWriter) : Helper to create HTTP responses
//	r (*http.Request) : HTTP request used to launch this function
//
// Examples:
//
//	a.updateOrder(w, &r)
//
// ===========================================================================================================
func (a *App) updateOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}
	var o oko.Order
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&o); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer r.Body.Close()
	o.ID = id

	if err := o.UpdateOrder(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, o)
}

// ===========================================================================================================
// Function called by DELETE HTTP route /order/x that aims at deleting an order
//
// Used on:
//
//	a (*App) : App struct containing the service necessary items
//
// Parameters:
//
//	w (http.ResponseWriter) : Helper to create HTTP responses
//	r (*http.Request) : HTTP request used to launch this function
//
// Examples:
//
//	a.deleteOrder(w, &r)
//
// ===========================================================================================================
func (a *App) deleteOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Order ID")
		return
	}

	checkOrderReq, _ := http.NewRequest("GET", "/order/"+strconv.Itoa(id), nil)
	checkOrderResp := executeRequest(checkOrderReq)

	if checkOrderResp.Code != 200 {
		respondWithError(w, http.StatusInternalServerError, "Unexpected order to delete.")
		return
	}

	o := oko.Order{ID: id}
	if err := o.DeleteOrder(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// ===========================================================================================================
// Initialize every HTTP route of our application
//
// Used on:
//
//	a (*App) : App struct containing the service necessary items
//
// ===========================================================================================================
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/orders", a.getOrders).Methods("GET")                 // Get information about all orders
	a.Router.HandleFunc("/order", a.createOrder).Methods("POST")               // Create an order and call sys order service
	a.Router.HandleFunc("/order/{id:[0-9]+}", a.getOrder).Methods("GET")       // Get information about an order
	a.Router.HandleFunc("/order/{id:[0-9]+}", a.updateOrder).Methods("PUT")    // Update an order
	a.Router.HandleFunc("/order/{id:[0-9]+}", a.deleteOrder).Methods("DELETE") // Delete an order
}
