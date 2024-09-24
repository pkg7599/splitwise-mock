package app

import (
	"encoding/json"
	"net/http"
	"os"
	"splitwise-api/internal"

	"github.com/gorilla/mux"
	"github.com/lpernett/godotenv"
)

type Api interface {
	Init() error
	Start() error
	Stop()
}

type ApiImpl struct {
	router *mux.Router

	exitChan chan bool
}

func CreateApp() (*ApiImpl, error) {
	if err := godotenv.Load("config.env"); err != nil {
		return nil, err
	}
	return &ApiImpl{
		exitChan: make(chan bool),
	}, nil
}

func (api *ApiImpl) Init() error {
	api.router = mux.NewRouter().StrictSlash(true)
	db, err := internal.PostgresClientInit(nil)

	// Migrate the DB schema
	internal.MigrateSchema(db)

	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}
	// Initialize routes
	api.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := "API RUNNING"
		status := 200
		resp := internal.SuccessResp(&status, &msg, map[string]string{"status": "HEALTHY"})
		json.NewEncoder(w).Encode(resp)
	}).Methods("GET")

	// Health Check
	api.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		msg := "API RUNNING"
		status := 200
		rqstContext := r.Context()
		db.SetContext(&rqstContext)

		if err := db.Ping(); err != nil {
			msg = err.Error()
			status = 500
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internal.ErrorResp(&status, &msg, map[string]string{"status": "DATABASE CONNECTION ERROR"}))
		}

		resp := internal.SuccessResp(&status, &msg, map[string]string{"status": "HEALTHY"})
		json.NewEncoder(w).Encode(resp)
	}).Methods("GET")

	// Initialize User Handler
	userHandler, err := internal.NewUserHandler()
	if err != nil {
		return err
	}
	// User Routes
	internal.UserRouter(api.router, *userHandler)
	
	// Initialize Expense Handler
	expenseHandler, err := internal.NewExpenseHandler()
	if err != nil {
		return err
	}
	// Expense Routes
	internal.ExpenseRouter(api.router, *expenseHandler)

	// Initialize Lender Handler
	lenderHandler, err := internal.NewLenderHandler()
	if err != nil {
		return err
	}
	// Expense Routes
	internal.LenderRouter(api.router, *lenderHandler)

	return nil
}

func (api *ApiImpl) Start() error {
	go func() {
		PORT := ":" + os.Getenv("APP_PORT")
		if PORT == ":" {
			PORT = ":8080"
		}
		// Start the server
		if err := http.ListenAndServe(PORT, api.router); err != nil {
			panic(err)
		}
	}()
	<-api.exitChan
	return nil
}

func (api *ApiImpl) Stop() {
	api.exitChan <- true
}
