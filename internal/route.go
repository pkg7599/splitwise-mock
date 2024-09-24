package internal

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	service *UserService
}

func NewUserHandler() (*UserHandler, error) {
	userService, err := UserServiceInit()
	if err != nil {
		return nil, err
	}
	return &UserHandler{service: userService}, nil
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	var statusCode int = http.StatusOK
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		statusCode = http.StatusBadRequest
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	if err := h.service.Add(user.Name, user.Email, user.PhoneNo); err != nil {
		statusCode = http.StatusInternalServerError
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
	}
	w.WriteHeader(statusCode)
	msg := "User added successfully"
	json.NewEncoder(w).Encode(SuccessResp(&statusCode, &msg, nil))
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	var params map[string]string = mux.Vars(r)
	var statusCode int = http.StatusOK
	w.Header().Set("Content-Type", "application/json")
	uidParsed, err := uuid.Parse(params["uid"])
	if err != nil {
		statusCode = http.StatusBadRequest
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	user, err := h.service.Get(uidParsed)
	if err != nil {
		statusCode = http.StatusInternalServerError
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(SuccessResp(&statusCode, nil, user))
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var params map[string]string = mux.Vars(r)
	var statusCode int = http.StatusOK
	w.Header().Set("Content-Type", "application/json")
	uidParsed, err := uuid.Parse(params["uid"])
	if err != nil {
		statusCode = http.StatusBadRequest
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	if err := h.service.Delete(uidParsed); err != nil {
		statusCode = http.StatusInternalServerError
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	w.WriteHeader(statusCode)
	msg := "User deleted successfully"
	json.NewEncoder(w).Encode(SuccessResp(&statusCode, &msg, nil))
}

type ExpenseHandler struct {
	service       *ExpenseService
	lenderService *LenderService
}

func NewExpenseHandler() (*ExpenseHandler, error) {
	expenseService, err := ExpenseServiceInit()
	if err != nil {
		return nil, err
	}
	lenderService, err := LenderServiceInit()
	if err != nil {
		return nil, err
	}
	return &ExpenseHandler{service: expenseService, lenderService: lenderService}, nil
}

func (es *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var expenseRequest ExpenseRequest
	var statusCode int = http.StatusOK
	if err := json.NewDecoder(r.Body).Decode(&expenseRequest); err != nil {
		statusCode = http.StatusBadRequest
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}

	if err := Validate(expenseRequest); err != nil {
		statusCode = http.StatusBadRequest
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	lenderId := expenseRequest.LenderId
	expenseType := expenseRequest.Type
	users := expenseRequest.Users
	amount := expenseRequest.Amount
	description := expenseRequest.Description
	var expenseBorrowers []*ExpenseBorrower
	if expenseType == "equal" {
		splitAmount := amount / float64(len(users))
		for _, userId := range users {
			if userId == lenderId {
				continue
			}
			expenseBorrowers = append(expenseBorrowers, &ExpenseBorrower{
				BorrowerId: userId,
				Amount:     splitAmount,
			})
		}
	} else if expenseType == "percent" {
		percents := expenseRequest.Percents
		for i, userId := range users {
			if userId == lenderId {
				continue
			}
			splitAmount := percents[i] * amount / 100
			expenseBorrowers = append(expenseBorrowers, &ExpenseBorrower{
				BorrowerId: userId,
				Amount:     splitAmount,
			})
		}
	} else if expenseType == "exact" {
		for i, userId := range users {
			expenseBorrowers = append(expenseBorrowers, &ExpenseBorrower{
				BorrowerId: userId,
				Amount:     expenseRequest.Values[i],
			})
		}
	} else {
		statusCode = http.StatusBadRequest
		errMsg := "error: expenseType should be in equal, percent, exact"
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	if err := es.service.Add(expenseType, amount, description, lenderId, expenseBorrowers); err != nil {
		statusCode = http.StatusInternalServerError
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	var lenders []*Lend
	for _, expenseBorrower := range expenseBorrowers {
		lenders = append(lenders, NewLender(lenderId, expenseBorrower.BorrowerId, expenseBorrower.Amount))
	}

	if err := Parallelize(es.lenderService.Upsert, lenders); err != nil {
		statusCode = http.StatusInternalServerError
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}

	w.WriteHeader(statusCode)
	msg := "Expense added successfully"
	json.NewEncoder(w).Encode(SuccessResp(&statusCode, &msg, nil))
}

func (es *ExpenseHandler) GetExpense(w http.ResponseWriter, r *http.Request) {
	var params map[string]string = mux.Vars(r)
	var statusCode int = http.StatusOK
	w.Header().Set("Content-Type", "application/json")
	uidParsed, err := ParseUUIDString(params["exId"])
	if err != nil {
		statusCode = http.StatusBadRequest
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	expense, err := es.service.Get(*uidParsed)
	if err != nil {
		statusCode = http.StatusInternalServerError
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(SuccessResp(&statusCode, nil, expense))
}

type LenderHandler struct {
	service *LenderService
}

func NewLenderHandler() (*LenderHandler, error) {
	lenderService, err := LenderServiceInit()
	if err != nil {
		return nil, err
	}
	return &LenderHandler{service: lenderService}, nil
}

func (lh *LenderHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	userId1 := queryParams.Get("userId1")
	userId2 := queryParams.Get("userId2")
	var statusCode int = http.StatusOK
	w.Header().Set("Content-Type", "application/json")
	userId1Parsed, err := ParseUUIDString(userId1)
	if err != nil {
		statusCode = http.StatusBadRequest
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	userId2Parsed, err := ParseUUIDString(userId2)
	if err != nil {
		statusCode = http.StatusBadRequest
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	lender, err := lh.service.GetBalance(*userId1Parsed, *userId2Parsed)
	if err != nil {
		statusCode = http.StatusInternalServerError
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(SuccessResp(&statusCode, nil, lender))
}

func (lh *LenderHandler) GetLendSummary(w http.ResponseWriter, r *http.Request) {
	var params map[string]string = mux.Vars(r)
	var statusCode int = http.StatusOK
	w.Header().Set("Content-Type", "application/json")
	uidParsed, err := ParseUUIDString(params["userId"])
	if err != nil {
		statusCode = http.StatusBadRequest
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	lenders, err := lh.service.GetLendSummary(*uidParsed)
	if err != nil {
		statusCode = http.StatusInternalServerError
		errMsg := err.Error()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResp(&statusCode, &errMsg, nil))
		return
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(SuccessResp(&statusCode, nil, lenders))
}
