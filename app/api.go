package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"splitwise-api/internal"
	"time"

	"github.com/gorilla/mux"
	"github.com/lpernett/godotenv"
)

var log internal.CustomLogger = internal.Log

type Api interface {
	Init() error
	Start() error
	Stop()
}

type ApiImpl struct {
	srv    *http.Server
	router *mux.Router
}

func CreateApp() (*ApiImpl, error) {
	if err := godotenv.Load("config.env"); err != nil {
		log.Error(fmt.Sprintf("error occurred in app create: %s", err))
		return nil, err
	}

	return &ApiImpl{
		router: mux.NewRouter().StrictSlash(true),
	}, nil
}

func (api *ApiImpl) SetupRoutes() {
	// Add User Routes
	userHandler, err := internal.NewUserHandler()
	if err != nil {
		log.Error(fmt.Sprintf("error occurred in user routes initialization: %s", err))
		panic(err)
	}
	internal.UserRouter(api.router, *userHandler)

	// Add Expense Routes
	expenseHandler, err := internal.NewExpenseHandler()
	if err != nil {
		log.Error(fmt.Sprintf("error occurred in expense routes initialization: %s", err))
		panic(err)
	}
	internal.ExpenseRouter(api.router, *expenseHandler)

	// Add Lender Routes
	lenderHandler, err := internal.NewLenderHandler()
	if err != nil {
		log.Error(fmt.Sprintf("error occurred in lender routes initialization: %s", err))
		panic(err)
	}
	internal.LenderRouter(api.router, *lenderHandler)
}

func (api *ApiImpl) Init() error {
	db, err := internal.PostgresClientInit(nil)

	// Migrate the DB schema
	internal.MigrateSchema(db)

	if err != nil {
		log.Error(fmt.Sprintf("error occurred in app initialization: %s", err))
		return err
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	if err := db.Ping(&ctx); err != nil {
		log.Error(fmt.Sprintf("error occurred in db connection: %s", err))
		return err
	}
	// Initialize routes
	api.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		msg := "API RUNNING"
		status := 200
		resp := internal.SuccessResp(&status, &msg, map[string]string{"status": "HEALTHY"})
		json.NewEncoder(w).Encode(resp)
	}).Methods("GET")

	// Health Check
	api.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		msg := "API RUNNING"
		status := 200
		rqstContext := r.Context()
		db.SetContext(&rqstContext)
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := db.Ping(&ctx); err != nil {
			msg = err.Error()
			log.Error(fmt.Sprintf("error occurred in health check: %s", msg))
			status = 500
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internal.ErrorResp(&status, &msg, map[string]string{"status": "DATABASE CONNECTION ERROR"}))
			return
		}

		resp := internal.SuccessResp(&status, &msg, map[string]string{"status": "HEALTHY"})
		json.NewEncoder(w).Encode(resp)
	}).Methods("GET")

	// Setup Routes for Services
	api.SetupRoutes()

	PORT := ":" + os.Getenv("APP_PORT")
	if PORT == ":" {
		PORT = ":8080"
	}

	// Initialize the server
	api.srv = &http.Server{
		Addr:    PORT,
		Handler: api.router,
	}
	return nil
}

func (api *ApiImpl) Start() error {
	go func() {
		// Start the server
		if err := api.srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error(fmt.Sprintf("error occurred in app start: %s", err))
			panic(err)
		}
	}()
	return nil
}

func (api *ApiImpl) Stop(t time.Duration) {
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	if err := api.srv.Shutdown(ctx); err != nil {
		log.Error(fmt.Sprintf("error occurred in app stop: %s", err))
		panic(err)
	}
}
