package internal

import (
	"context"
	"fmt"
)

type IDao[T any] interface {
	Client(*context.Context) IClient
	Create(*context.Context, interface{}) error
	Update(*context.Context, T) error
	Delete(*context.Context, *T) error
	Read(*context.Context, map[string]interface{}) ([]T, error)
}

type Dao[T any] struct {
	dbClient IClient
	ctx      *context.Context
}

func DaoInit[T any](ctx *context.Context) (IDao[T], error) {
	var context *context.Context

	if ctx != nil {
		context = ctx
	} else {
		context = ctx
	}

	// init db client
	dbClient, err := PostgresClientInit(ctx)
	if err != nil {
		Log.Error(fmt.Sprintf("dao initialization error: %s", err.Error()))
		return nil, err
	}
	return &Dao[T]{
		dbClient: dbClient,
		ctx:      context,
	}, nil
}

func (dao *Dao[T]) Client(ctx *context.Context) IClient {
	return dao.dbClient
}

// Create a new entity
// @param ctx *context.Context: Context
// @param entity interface{}: The entity to create
// @return error: The error if any
func (dao *Dao[T]) Create(ctx *context.Context, entity interface{}) error {
	dao.dbClient.StartSession(ctx)
	result := dao.dbClient.DbClient(ctx).Create(entity)
	dao.dbClient.CommitSession()
	if result.Error != nil {
		Log.Info(fmt.Sprintf("entity: %+v created", entity))
	}
	return result.Error
}

// Update entity
// @param ctx *context.Context: Context
// @param entity interface{}: The entity to update
// @return error: The error if any
func (dao *Dao[T]) Update(ctx *context.Context, entity T) error {
	dao.dbClient.StartSession(ctx)
	result := dao.dbClient.DbClient(ctx).Save(entity)
	dao.dbClient.CommitSession()
	if result.Error != nil {
		Log.Info(fmt.Sprintf("entity: %+v updated", entity))
	}
	return result.Error
}

// Delete entity
// @param ctx *context.Context: Context
// @param entity interface{}: The entity to delete
// @return error: The error if any
func (dao *Dao[T]) Delete(ctx *context.Context, entity *T) error {
	dao.dbClient.StartSession(ctx)
	result := dao.dbClient.DbClient(ctx).Delete(entity)
	dao.dbClient.CommitSession()
	if result.Error != nil {
		Log.Info(fmt.Sprintf("entity: %+v deleted", entity))
	}
	return result.Error
}

// Read entities
// @param ctx *context.Context: Context
// @param filter map[string]interface{}: The filter to apply
// @return []T: Search Result
// @return error: The error if any
func (dao *Dao[T]) Read(ctx *context.Context, filter map[string]interface{}) ([]T, error) {
	var results []T
	resp := dao.dbClient.DbClient(ctx).Where(filter).Find(&results)
	if resp.Error != nil {
		Log.Info(fmt.Sprintf("records Found: %d", len(results)))
	}
	return results, resp.Error
}
