package internal

import (
	"context"
	"errors"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type IClient interface {
	DbClient(*context.Context) *gorm.DB
	SetContext(*context.Context)
	ResetContext()
	StartSession(*context.Context) error
	CommitSession() error
	AbortSession() error
	Ping(*context.Context) error
}

type PostgresClient struct {
	Client  *gorm.DB
	session *gorm.DB
	ctx     *context.Context
}

func PostgresClientInit(ctx *context.Context) (IClient, error) {

	POSTGRES_DSN := os.Getenv("POSTGRES_DSN")
	DB_NAME := os.Getenv("DB_NAME")

	if POSTGRES_DSN == "" {
		Log.Error("env not found: POSTGRES_DSN")
		return nil, errors.New("POSTGRES_DSN is not set")
	}

	if DB_NAME == "" {
		Log.Error("env not found: DB_NAME")
		return nil, errors.New("DB_NAME is not set")
	}

	dsn := fmt.Sprintf(POSTGRES_DSN, DB_NAME)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: GormLogger()})
	if err != nil {
		Log.Error(fmt.Sprintf("postgres client init error: %s", err.Error()))
		return nil, err
	}
	if ctx == nil {
		context := context.TODO()
		ctx = &context
	}
	client := db.WithContext(*ctx)
	return &PostgresClient{
		Client:  client,
		session: client.Begin(),
		ctx:     ctx,
	}, nil
}

func (c *PostgresClient) DbClient(ctx *context.Context) *gorm.DB {
	if ctx != nil {
		c.SetContext(ctx)
	}
	return c.Client
}

func (c *PostgresClient) SetContext(ctx *context.Context) {
	if ctx == nil {
		context := context.TODO()
		ctx = &context
	}
	Log.Info("Setting new context for PostgresClient")
	c.ctx = ctx
	c.Client = c.Client.WithContext(*c.ctx)
}

func (c *PostgresClient) ResetContext() {
	Log.Info("Resetting context for PostgresClient")
	ctx := context.TODO()
	c.ctx = &ctx
}

func (c *PostgresClient) StartSession(ctx *context.Context) error {
	c.SetContext(ctx)
	c.session = c.Client.WithContext(*c.ctx).Begin()
	Log.Info("Starting new session for PostgresClient")
	return c.session.Error
}

func (c *PostgresClient) CommitSession() error {
	db := c.session.Commit()
	Log.Info("Committing session for PostgresClient")
	return db.Error
}

func (c *PostgresClient) AbortSession() error {
	db := c.session.Rollback()
	Log.Info("Aborting session for PostgresClient")
	return db.Error
}

func (c *PostgresClient) Ping(ctx *context.Context) error {
	db, err := c.Client.DB()
	c.SetContext(ctx)
	if err != nil {
		Log.Error(fmt.Sprintf("postgres client ping error: %s", err.Error()))
		return err
	}
	return db.PingContext(*c.ctx)
}

// Auto Initialize the schema into DB
func MigrateSchema(c IClient) error {

	schemas := []interface{}{
		&User{},
		&Lend{},
		&Expense{},
		&ExpenseBorrower{},
	}

	for _, schema := range schemas {
		err := c.DbClient(nil).AutoMigrate(schema)
		if err != nil {
			Log.Error(fmt.Sprintf("postgres client migrate schema error: %s", err.Error()))
			return err
		}
	}
	return nil
}
