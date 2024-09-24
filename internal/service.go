package internal

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

type LenderService struct {
	dao IDao[Lend]
}

func LenderServiceInit() (*LenderService, error) {
	dao, err := DaoInit[Lend](nil)
	if err != nil {
		return nil, err
	}
	return &LenderService{dao: dao}, nil
}

func (ls *LenderService) Add(borrowerId uuid.UUID, lenderId uuid.UUID, amount float64) error {
	var lend *Lend = NewLender(lenderId, borrowerId, amount)
	return ls.dao.Create(&lend)
}

func (ls *LenderService) GetBalance(userId1 uuid.UUID, userId2 uuid.UUID) (*Lend, error) {
	var lend Lend
	lId := GenerateUUIDFromUUIDs(userId1, userId2)
	dbClient := ls.dao.Client()
	dbClient.StartSession()
	lenders, err := ls.dao.Read(map[string]interface{}{"l_id": lId})
	if err != nil {
		return nil, err
	}
	if len(lenders) > 0 {
		lend = lenders[0]
	}
	return &lend, nil
}

func (ls *LenderService) GetLendSummary(userId uuid.UUID) ([]*Lend, error) {
	var lends []*Lend
	dbClient := ls.dao.Client()
	resp := dbClient.DbClient().Where("lender_id = ? OR borrower_id = ?", userId, userId).Find(&lends)
	return lends, resp.Error
}

func (ls *LenderService) Upsert(lend *Lend) error {
	dbClient := ls.dao.Client()
	conflictField, err := GetDbFieldName("LId", lend)
	if err != nil {
		return err
	}
	updateField, err := GetDbFieldName("Amount", lend)
	if err != nil {
		return err
	}

	lenderIdName, err := GetDbFieldName("LenderId", lend)
	if err != nil {
		return err
	}

	dbClient.StartSession()
	resp := dbClient.DbClient().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: conflictField}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			updateField: clause.Expr{
				SQL:  fmt.Sprintf("CASE WHEN lends.%s = ? THEN lends.%s + ? WHEN lends.%s = ? THEN lends.%s - ? END", lenderIdName, updateField, lenderIdName, updateField),
				Vars: []interface{}{lend.LenderId, lend.Amount, lend.BorrowerId, lend.Amount},
			},
		}),
	}).Create(lend)
	return resp.Error
}

type UserService struct {
	dao IDao[User]
}

func UserServiceInit() (*UserService, error) {
	dao, err := DaoInit[User](nil)
	if err != nil {
		return nil, err
	}
	return &UserService{dao: dao}, nil
}

func (us *UserService) Add(name string, email string, phone string) error {
	var user *User = NewUser(name, email, phone)
	if err := us.dao.Create(&user); err != nil {
		return err
	}

	return nil
}

func (us *UserService) Get(id uuid.UUID) (*User, error) {
	user, err := us.dao.Read(map[string]interface{}{"uid": id})
	if err != nil {
		return nil, err
	}
	return &user[0], nil
}

func (us *UserService) Delete(id uuid.UUID) error {
	user := User{UId: id}
	if err := us.dao.Delete(&user); err != nil {
		return err
	}
	return nil
}

type ExpenseService struct {
	dao IDao[Expense]
}

func ExpenseServiceInit() (*ExpenseService, error) {
	dao, err := DaoInit[Expense](nil)
	if err != nil {
		return nil, err
	}
	return &ExpenseService{dao: dao}, nil
}

func (es *ExpenseService) Add(
	category string, amount float64, description string, lenderId uuid.UUID, expenseBorrowers []*ExpenseBorrower,
) error {
	var expense *Expense = NewExpense(category, amount, description, lenderId, expenseBorrowers)
	if err := es.dao.Create(&expense); err != nil {
		return err
	}

	return nil
}

func (es *ExpenseService) Get(id uuid.UUID) (*Expense, error) {
	expense, err := es.dao.Read(map[string]interface{}{"ex_id": id})
	if err != nil {
		return nil, err
	}
	return &expense[0], nil
}
