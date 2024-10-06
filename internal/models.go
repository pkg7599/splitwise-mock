package internal

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// User Model
type User struct {
	UId       uuid.UUID `json:"uId,omitempty" gorm:"primaryKey;type:uuid"`
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email,omitempty"`
	PhoneNo   string    `json:"phoneNo,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

func NewUser(name string, email string, phoneNo string) *User {
	return &User{
		UId:       GenerateUUIdV6(),
		Name:      name,
		Email:     email,
		PhoneNo:   phoneNo,
		CreatedAt: time.Now().UTC(),
	}
}

// Lend Model
type Lend struct {
	LId        uuid.UUID `json:"lId,omitempty" gorm:"primaryKey;type:uuid"`
	LenderId   uuid.UUID `json:"lenderId,omitempty" gorm:"type:uuid"`
	Lender     User      `json:"-" gorm:"foreignKey:LenderId"`
	BorrowerId uuid.UUID `json:"borrowerId,omitempty" gorm:"type:uuid"`
	Borrower   User      `json:"-" gorm:"foreignKey:BorrowerId"`
	Amount     float64   `default:"0" json:"amount"`
	UpdatedAt  time.Time `json:"updatedAt,omitempty"`
}

func NewLender(lenderId uuid.UUID, borrowerId uuid.UUID, amount float64) *Lend {
	return &Lend{
		LId:        GenerateUUIDFromUUIDs(lenderId, borrowerId),
		LenderId:   lenderId,
		BorrowerId: borrowerId,
		Amount:     amount,
		UpdatedAt:  time.Now().UTC(),
	}
}

type ExpenseRequest struct {
	Type        string      `json:"type,omitempty" validate:"required"`
	LenderId    uuid.UUID   `json:"lenderId,omitempty" validate:"required"`
	Amount      float64     `json:"amount,omitempty" validate:"required,gt=0"`
	Description string      `json:"description,omitempty"`
	Users       []uuid.UUID `json:"users,omitempty" validate:"required"`
	Percents    []float64   `json:"percents,omitempty" validate:"required_if=Type percent"`
	Values      []float64   `json:"values,omitempty" validate:"required_if=Type exact"`
}

func Validate(expenseRequest ExpenseRequest) error {
	if validationErr := validator.New().Struct(expenseRequest); validationErr != nil {
		return validationErr
	}
	expenseType := expenseRequest.Type
	if expenseType != "equal" && expenseType != "exact" && expenseType != "percent" {
		return fmt.Errorf("invalid expense type")
	}
	if len(expenseRequest.Users) < 2 && expenseType == "equal" {
		return fmt.Errorf("at least 2 users are required to split equally")
	} else if len(expenseRequest.Users) < 1 && expenseType == "exact" {
		return fmt.Errorf("at least 1 user is required to split exactly")
	} else if len(expenseRequest.Users) < 1 && expenseType == "percent" {
		return fmt.Errorf("at least 1 user are required to split by percent")
	}
	if expenseType == "percent" {
		sum := 0.0
		for _, val := range expenseRequest.Percents {
			if val < 0 || val > 100 {
				return fmt.Errorf("invalid percent value")
			}
			sum += val
		}
		if sum != 100.0 {
			return fmt.Errorf("validationError: summation of percents should be 100")
		}
	}
	if expenseType == "exact" {
		sum := 0.0
		for _, val := range expenseRequest.Values {
			sum += val
		}
		if sum != expenseRequest.Amount {
			return fmt.Errorf("validationError: summation of values should be equal to amount lended")
		}
	}
	return nil
}

type ExpenseBorrower struct {
	ExpenseId  uuid.UUID `json:"expenseId,omitempty" gorm:"primaryKey;type:uuid"`
	BorrowerId uuid.UUID `json:"borrowerId,omitempty" gorm:"primaryKey;type:uuid"`
	Borrower   User      `json:"-" gorm:"foreignKey:BorrowerId"`
	Amount     float64   `json:"amount,omitempty"`
	IsPaid     bool      `json:"isPaid,omitempty" gorm:"default:false"`
}

func NewExpenseBorrower(expenseId uuid.UUID, borrowerId uuid.UUID, amount float64) *ExpenseBorrower {
	return &ExpenseBorrower{
		ExpenseId:  expenseId,
		BorrowerId: borrowerId,
		Amount:     amount,
	}
}

// Expense Model
type Expense struct {
	ExId             uuid.UUID          `json:"exId,omitempty" gorm:"primaryKey;type:uuid"`
	Category         string             `json:"category,omitempty"`
	Amount           float64            `json:"amount,omitempty"`
	Description      string             `json:"description,omitempty"`
	CreatedAt        time.Time          `json:"createdAt,omitempty"`
	LenderId         uuid.UUID          `json:"lenderId,omitempty" gorm:"type:uuid"`
	Lender           User               `json:"-" gorm:"foreignKey:LenderId"`
	ExpenseBorrowers []*ExpenseBorrower `json:"borrowers,omitempty" gorm:"foreignKey:ExpenseId"`
}

func NewExpense(category string, amount float64, description string, lenderId uuid.UUID, expenseBorrowers []*ExpenseBorrower) *Expense {
	exId := GenerateUUIdV6()
	for _, expBorrower := range expenseBorrowers {
		expBorrower.ExpenseId = exId
	}
	return &Expense{
		ExId:             exId,
		Category:         category,
		Amount:           amount,
		Description:      description,
		CreatedAt:        time.Now().UTC(),
		LenderId:         lenderId,
		ExpenseBorrowers: expenseBorrowers,
	}
}

// API Response Model
type Response struct {
	Timestamp time.Time   `json:"timestamp"`
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}

func SuccessResp(status *int, msg *string, data interface{}) *Response {
	var status_code int = 200
	var message string = "Success"

	if status != nil {
		status_code = *status
	}

	if msg != nil {
		message = *msg
	}

	return &Response{
		Timestamp: time.Now().UTC(),
		Status:    status_code,
		Message:   message,
		Data:      data,
	}
}

func ErrorResp(status *int, msg *string, data interface{}) *Response {
	var status_code int = 500
	var message string = "INTERNAL SERVER ERROR"

	if status != nil {
		status_code = *status
	}

	if msg != nil {
		message = *msg
	}

	return &Response{
		Timestamp: time.Now().UTC(),
		Status:    status_code,
		Message:   message,
		Data:      data,
	}
}
