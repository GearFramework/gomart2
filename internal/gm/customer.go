package gm

import (
	"context"
	"fmt"
	"github.com/GearFramework/gomart/internal/gm/types"
	"golang.org/x/crypto/bcrypt"
)

var (
	sqlInsertCustomer = `
		INSERT INTO gomartspace.customers 
		       (login, password)
		VALUES ($1, $2)
	 RETURNING id
	`
	sqlGetCustomerByLogin = `
		SELECT id,
		       login,
		       password,
		       balance,
		       withdraw
		  FROM gomartspace.customers
		 WHERE login = $1
	`
	sqlGetCustomerByID = `
		SELECT id,
		       login,
		       password,
		       balance,
		       withdraw
		  FROM gomartspace.customers
		 WHERE id = $1
	`
	sqlUpdateCustomerBalance = `
		UPDATE gomartspace.customers
		   SET balance = balance + $2
		 WHERE id = $1
	`
)

type Customer struct {
	Id       int64   `db:"id" json:"id"`
	Login    string  `db:"login" json:"login"`
	Password string  `db:"password" json:"-"`
	Balance  float32 `db:"balance" json:"balance"`
	Withdraw float32 `db:"withdraw" json:"withdrawn"`
}

// CreateCustomer создание клиента
func (gm *GopherMartApp) CreateCustomer(data types.CustomerRegisterRequest) (int64, error) {
	hash, err := createHashPassword(data.Password)
	if err != nil {
		return 0, err
	}
	return gm.insertCustomer(data.GetCtx(), data.Login, hash)
}

func (gm *GopherMartApp) insertCustomer(ctx context.Context, login, hashedPassword string) (int64, error) {
	row, err := gm.Storage.Insert(ctx, sqlInsertCustomer, login, hashedPassword)
	if err != nil {
		gm.logger.Warn(err.Error())
		return 0, types.ErrCustomerAlreadyExists
	}
	var userID int64
	err = row.Scan(&userID)
	if err != nil {
		gm.logger.Warn(err.Error())
		return 0, types.ErrRegistration
	}
	gm.logger.Infof("new customer ID: %d", userID)
	return userID, nil
}

func createHashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func (gm *GopherMartApp) GetCustomer(data types.CustomerLoginRequest) (*Customer, error) {
	var customer Customer
	err := gm.Storage.Get(data.GetCtx(), &customer, sqlGetCustomerByLogin, data.Login)
	if err != nil {
		gm.logger.Error(err.Error())
		return nil, types.ErrCustomerNotFound
	}
	if !checkHashPassword(customer.Password, data.Password) {
		return nil, types.ErrCustomerLogin
	}
	return &customer, nil
}

func checkHashPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (gm *GopherMartApp) GetCustomerByID(ctx context.Context, customerID int64) (*Customer, error) {
	var customer Customer
	err := gm.Storage.Get(ctx, &customer, sqlGetCustomerByID, customerID)
	return &customer, err
}

func (gm *GopherMartApp) UpdateCustomerBalance(ctx context.Context, customer *Customer, appendBalance float32) (float32, error) {
	if err := gm.Storage.Update(ctx, sqlUpdateCustomerBalance, customer.Id, appendBalance); err != nil {
		return customer.Balance, err
	}
	fmt.Println(customer.Balance)
	customer.Balance += appendBalance
	return customer.Balance, nil
}
