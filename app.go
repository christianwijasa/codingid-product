package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user string, password string, dbName string, dbConnection string) {
	fmt.Println("Start to Initialize App")

	var err error

	connectionString := user + ":" + password + "@/" + dbName

	a.DB, err = sql.Open(dbConnection, connectionString)
	if err != nil {
		log.Fatalf("[sql.Open] %s", err.Error())
	}

	driver, err := mysql.WithInstance(a.DB, &mysql.Config{})
	if err != nil {
		log.Fatalf("[mysql.WithInstance] %s", err.Error())
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./database/migrations",
		dbName,
		driver,
	)
	if err != nil {
		log.Fatalf("[NewWithDatabaseInstance] %s", err.Error())
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("[Up] %s", err.Error())
	}

	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

func (a *App) Run(port string) {
	fmt.Println("App is running on port", port)

	if err := http.ListenAndServe(":"+port, a.Router); err != nil {
		log.Fatalf("[ListendAndServe] %s", err.Error())
	}
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/products", a.getProducts).Methods(http.MethodGet)
	a.Router.HandleFunc("/product/{sku}", a.getProductBySKU).Methods(http.MethodGet)
	a.Router.HandleFunc("/product", a.createProduct).Methods(http.MethodPost)
	a.Router.HandleFunc("/product/{sku}", a.deleteProduct).Methods(http.MethodDelete)
}

func (a *App) getProducts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.FormValue("limit"))
	offset, _ := strconv.Atoi(r.FormValue("offset"))

	if limit <= 0 {
		limit = 10
	}

	products, err := getProducts(a.DB, limit, offset)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Product not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if len(products) == 0 {
		respondWithError(w, http.StatusNotFound, "Product not found")
		return
	}

	var respHTTP = struct {
		Products []product `json:"products"`
	}{
		Products: products,
	}

	respondWithJSON(w, http.StatusOK, respHTTP)
}

func (a *App) getProductBySKU(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p := product{SKU: vars["sku"]}

	if err := p.getProductBySKU(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Product not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	var respHTTP = struct {
		Product product `json:"product"`
	}{
		Product: p,
	}

	respondWithJSON(w, http.StatusOK, respHTTP)
}

func (a *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var p product

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := p.createProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	p := product{SKU: vars["sku"]}
	if err := p.deleteProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	fmt.Println(message)
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	fmt.Println(string(response))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
