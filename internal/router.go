package internal

import "github.com/gorilla/mux"

func UserRouter(r *mux.Router, handler UserHandler) {
	userRoute := r.PathPrefix("/user").Subrouter()
	userRoute.HandleFunc("", handler.CreateUser).Methods("POST")
	userRoute.HandleFunc("/{uid}", handler.GetUser).Methods("GET")
	userRoute.HandleFunc("/{uid}", handler.DeleteUser).Methods("DELETE")
}

func ExpenseRouter(r *mux.Router, handler ExpenseHandler) {
	expenseRoute := r.PathPrefix("/expense").Subrouter()
	expenseRoute.HandleFunc("", handler.CreateExpense).Methods("POST")
	expenseRoute.HandleFunc("/{exId}", handler.GetExpense).Methods("GET")
}

func LenderRouter(r *mux.Router, handler LenderHandler) {
	lenderRoute := r.PathPrefix("/lender").Subrouter()
	lenderRoute.HandleFunc("/", handler.GetBalance).Methods("GET")
	lenderRoute.HandleFunc("/{userId}", handler.GetLendSummary).Methods("GET")
}
