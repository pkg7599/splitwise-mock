package internal

import (
	"context"
	"errors"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type IClient interface {
	DbClient() *gorm.DB
	SetContext(*context.Context)
	ResetContext()
	StartSession() error
	CommitSession() error
	AbortSession() error
	Ping() error
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
		return nil, errors.New("POSTGRES_DSN is not set")
	}

	if DB_NAME == "" {
		return nil, errors.New("DB_NAME is not set")
	}

	dsn := fmt.Sprintf(POSTGRES_DSN, DB_NAME)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info)})

	if err != nil {
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

func (c *PostgresClient) DbClient() *gorm.DB {
	return c.Client
}

func (c *PostgresClient) SetContext(ctx *context.Context) {
	c.ctx = ctx
}

func (c *PostgresClient) ResetContext() {
	ctx := context.TODO()
	c.ctx = &ctx
}

func (c *PostgresClient) StartSession() error {
	c.session = c.Client.WithContext(*c.ctx).Begin()
	return c.session.Error
}

func (c *PostgresClient) CommitSession() error {
	db := c.session.Commit()
	return db.Error
}

func (c *PostgresClient) AbortSession() error {
	db := c.session.Rollback()
	return db.Error
}

func (c *PostgresClient) Ping() error {
	db, err := c.Client.DB()
	if err != nil {
		return err
	}
	return db.Ping()
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
		err := c.DbClient().AutoMigrate(schema)
		if err != nil {
			return err
		}
	}
	return nil
}
