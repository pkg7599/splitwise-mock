package internal

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

type LenderService struct {
	dao IDao[Lend]
}

func LenderServiceInit() (*LenderService, error) {
	Log.Info("lender service init...")
	dao, err := DaoInit[Lend](nil)
	if err != nil {
		Log.Error(fmt.Sprintf("lender service init error: %s", err.Error()))
		return nil, err
	}
	return &LenderService{dao: dao}, nil
}

func (ls *LenderService) Add(ctx *context.Context, borrowerId uuid.UUID, lenderId uuid.UUID, amount float64) error {
	var lend *Lend = NewLender(lenderId, borrowerId, amount)
	return ls.dao.Create(ctx, &lend)
}

func (ls *LenderService) GetBalance(ctx *context.Context, userId1 uuid.UUID, userId2 uuid.UUID) (*Lend, error) {
	var lend Lend
	lId := GenerateUUIDFromUUIDs(userId1, userId2)
	dbClient := ls.dao.Client(ctx)
	dbClient.StartSession(ctx)
	lIdName, err := GetDbFieldName("LId", lend)
	if err != nil {
		return nil, err
	}
	lenders, err := ls.dao.Read(ctx, map[string]interface{}{lIdName: lId})
	if err != nil {
		Log.Error(fmt.Sprintf("get balance error: %s", err.Error()))
		return nil, err
	}
	if len(lenders) > 0 {
		lend = lenders[0]
	}
	return &lend, nil
}

func (ls *LenderService) GetLendSummary(ctx *context.Context, userId uuid.UUID) ([]*Lend, error) {
	var lends []*Lend
	dbClient := ls.dao.Client(ctx)
	lenderIdFieldName, err := GetDbFieldName("LenderId", lends)
	if err != nil {
		return nil, err
	}
	borrowerIdFieldName, err := GetDbFieldName("BorrowerId", lends)
	if err != nil {
		return nil, err
	}
	resp := dbClient.DbClient(ctx).Where("? = ? OR ? = ?", lenderIdFieldName, borrowerIdFieldName, userId, userId).Find(&lends)
	return lends, resp.Error
}

func (ls *LenderService) UpdatePayment(ctx *context.Context, lenderId uuid.UUID, borrowerId uuid.UUID, amount float64) error {
	es, err := ExpenseServiceInit()
	if err != nil {
		return err
	}
	lend, err := ls.GetBalance(ctx, lenderId, borrowerId)
	if err != nil {
		return err
	}
	if lend.Amount != amount && lend.LenderId == lenderId {
		return fmt.Errorf("amount mismatch error: amount due: %f", lend.Amount)
	} else if lend.Amount != -amount && lend.LenderId == borrowerId {
		return fmt.Errorf("amount mismatch error: amount due: %f", -lend.Amount)
	}
	dbClient := ls.dao.Client(ctx)
	dbClient.StartSession(ctx)
	resp := dbClient.DbClient(ctx).Model(lend).Update("amount", 0)
	if resp.Error != nil {
		dbClient.AbortSession()
		return resp.Error
	}
	err = es.UpdatePayment(ctx, lenderId, borrowerId)
	if err != nil {
		dbClient.AbortSession()
		return err
	}
	dbClient.CommitSession()
	return nil
}

func (ls *LenderService) Upsert(ctx *context.Context, lend *Lend) error {
	dbClient := ls.dao.Client(ctx)
	conflictField, err := GetDbFieldName("LId", lend)
	if err != nil {
		Log.Error(fmt.Sprintf("Upsert error: %s", err.Error()))
		return err
	}
	updateField, err := GetDbFieldName("Amount", lend)
	if err != nil {
		Log.Error(fmt.Sprintf("get db field name error: %s", err.Error()))
		return err
	}

	lenderIdName, err := GetDbFieldName("LenderId", lend)
	if err != nil {
		Log.Error(fmt.Sprintf("get db field name error: %s", err.Error()))
		return err
	}

	dbClient.StartSession(ctx)
	resp := dbClient.DbClient(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: conflictField}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			updateField: clause.Expr{
				SQL:  fmt.Sprintf("CASE WHEN lends.%s = ? THEN lends.%s + ? WHEN lends.%s = ? THEN lends.%s - ? END", lenderIdName, updateField, lenderIdName, updateField),
				Vars: []interface{}{lend.LenderId, lend.Amount, lend.BorrowerId, lend.Amount},
			},
		}),
	}).Create(lend)
	if resp.Error != nil {
		Log.Info(fmt.Sprintf("upserted: %+v", lend))
	}
	return resp.Error
}

type UserService struct {
	dao IDao[User]
}

func UserServiceInit() (*UserService, error) {
	Log.Info("user service init...")
	dao, err := DaoInit[User](nil)
	if err != nil {
		Log.Error(fmt.Sprintf("user service init error: %s", err.Error()))
		return nil, err
	}
	return &UserService{dao: dao}, nil
}

func (us *UserService) Add(ctx *context.Context, name string, email string, phone string) error {
	var user *User = NewUser(name, email, phone)
	if err := us.dao.Create(ctx, &user); err != nil {
		return err
	}

	return nil
}

func (us *UserService) Get(ctx *context.Context, id uuid.UUID) (*User, error) {
	user, err := us.dao.Read(ctx, map[string]interface{}{"uid": id})
	if err != nil {
		return nil, err
	}
	return &user[0], nil
}

func (us *UserService) Delete(ctx *context.Context, id uuid.UUID) error {
	user := User{UId: id}
	if err := us.dao.Delete(ctx, &user); err != nil {
		return err
	}
	return nil
}

type ExpenseService struct {
	dao IDao[Expense]
}

func ExpenseServiceInit() (*ExpenseService, error) {
	Log.Info("expense service init...")
	dao, err := DaoInit[Expense](nil)
	if err != nil {
		Log.Error(fmt.Sprintf("expense service init error: %s", err.Error()))
		return nil, err
	}
	return &ExpenseService{dao: dao}, nil
}

func (es *ExpenseService) Add(
	ctx *context.Context, category string, amount float64, description string, lenderId uuid.UUID, expenseBorrowers []*ExpenseBorrower,
) error {
	var expense *Expense = NewExpense(category, amount, description, lenderId, expenseBorrowers)
	if err := es.dao.Create(ctx, &expense); err != nil {
		return err
	}

	return nil
}

func (es *ExpenseService) Get(ctx *context.Context, id uuid.UUID) (*Expense, error) {
	expense, err := es.dao.Read(ctx, map[string]interface{}{"ex_id": id})
	if err != nil {
		return nil, err
	}
	return &expense[0], nil
}

func (es *ExpenseService) UpdatePayment(ctx *context.Context, lenderId uuid.UUID, borrowerId uuid.UUID) error {
	var expenseBorrowers []ExpenseBorrower
	dbClient := es.dao.Client(ctx).DbClient(ctx)
	lenderIdFieldName, err := GetDbFieldName("LenderId", Lend{})
	if err != nil {
		return err
	}
	expenseIdFieldName, err := GetDbFieldName("ExpenseId", expenseBorrowers)
	if err != nil {
		return err
	}
	exIdFieldName, err := GetDbFieldName("ExId", Expense{})
	if err != nil {
		return err
	}
	borrowerIdFieldName, err := GetDbFieldName("BorrowerId", expenseBorrowers)
	if err != nil {
		return err
	}
	resp := dbClient.Model(&expenseBorrowers).Joins("JOIN expenses ON expense_borrowers." + expenseIdFieldName + " = expenses." + exIdFieldName).Where(
		map[string]interface{}{
			"expenses." + lenderIdFieldName:            lenderId,
			"expense_borrowers." + borrowerIdFieldName: borrowerId,
		},
	).Find(&expenseBorrowers)
	if resp.Error != nil {
		return resp.Error
	}

	var expenseIds []uuid.UUID
	for expenseBorrower := range expenseBorrowers {
		expenseIds = append(expenseIds, expenseBorrowers[expenseBorrower].ExpenseId)
	}
	isPaidFieldName, err := GetDbFieldName("IsPaid", expenseBorrowers)
	if err != nil {
		return err
	}
	resp = dbClient.Model(ExpenseBorrower{}).Where(expenseIdFieldName+" IN ? AND "+borrowerIdFieldName+" = ?", expenseIds, borrowerId).Update(isPaidFieldName, true)
	return resp.Error
}
