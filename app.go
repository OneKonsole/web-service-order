package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"

	"encoding/json"
	"net/http"
	"strconv"

	oko "github.com/OneKonsole/order-model"

	"github.com/gorilla/mux"

	_ "github.com/lib/pq"

	"github.com/go-playground/validator/v10"
)

type App struct {
	Router    *mux.Router
	DB        *sql.DB
	Validator *validator.Validate
	AppConf   *AppConf
}

type AppConf struct {
	ServedPort    string `json:"served_port"`     // e.g. "8010"
	DBUser        string `json:"db_user"`         // e.g. "MyUsername"
	DBPassword    string `json:"db_password"`     // e.g. "MyPassword1!"
	DBDestination string `json:"db_URL"`          // e.g. "localhost" || "myservice.mynamespace.svc.cluster.local" || "onekonsole.fr"
	DBName        string `json:"db_name"`         // e.g. "order"
	SysServiceUrl string `json:"sys_service_url"` // e.g. "http://localhost:8020/sys-service/"
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
func (a *App) Initialize() {

	fmt.Print("\nInitializing app...\n")

	connectionString := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		a.AppConf.DBDestination, 5432, a.AppConf.DBUser, a.AppConf.DBPassword, a.AppConf.DBName)

	fmt.Print(connectionString)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}

	err = a.DB.Ping()
	fmt.Printf("\nTESSSSSSSSSSST %s\n", err)

	a.Router = mux.NewRouter()

	// Helper to validate user inputs concerning orders management
	a.Validator = validator.New()
	a.Validator.RegisterValidation("isvalidclustername", isValidClusterName)
	a.Validator.RegisterValidation("startswithalphanum", startsWithAlphanum)
	a.Validator.RegisterValidation("endswithalphanum", endWithAlphanum)
	a.Validator.RegisterValidation("uuid", isUUID)

	a.initializeRoutes()
}

func (appConf *AppConf) Initialize() {
	appConf.ServedPort = os.Getenv("served_port")
	appConf.DBUser = os.Getenv("db_user")
	appConf.DBPassword = os.Getenv("db_password")
	appConf.DBDestination = os.Getenv("db_URL")
	appConf.DBName = os.Getenv("db_name")
	appConf.SysServiceUrl = os.Getenv("sys_service_url")

	fmt.Printf("Appconf: %s\n", appConf)

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
func (a *App) Run() {
	log.Fatal(http.ListenAndServe(":"+a.AppConf.ServedPort, a.Router))
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

	if err := decoder.Decode(&o); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := a.Validator.Struct(o); err != nil {
		respondWithError(w, http.StatusBadRequest, "One or more parameters do not match the required format")
		return
	}

	if err := o.CreateOrder(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Encode order as a HTTP Reader (io.Reader) in order to make request
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(o)

	// Contact sys-order service
	resp, err := http.Post(a.AppConf.SysServiceUrl, "application/json", buf)
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

	if err := a.Validator.Struct(o); err != nil {
		respondWithError(w, http.StatusBadRequest, "One or more parameters do not match the required format")
		return
	}

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
