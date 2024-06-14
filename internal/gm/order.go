package gm

import (
	"context"
	"fmt"
	"github.com/GearFramework/gomart/internal/gm/types"
	"time"
)

var (
	sqlInsertOrder = `
		INSERT INTO gomartspace.orders
			   (number, customer_id, accrual)
		VALUES ($1, $2, $3)
	`
	sqlGetOrderByID = `
		SELECT number,
		       customer_id,
			   accrual
		  FROM gomartspace.orders
		 WHERE number = $1
	`
	sqlGetCustomerOrders = `
		SELECT number,
		       uploaded_at,
		       status,
		       accrual
		  FROM gomartspace.orders
		 WHERE customer_id = $1
		 ORDER BY uploaded_at DESC
	`
)

func (gm *GopherMartApp) NewOrder(number string, customerID int64, status string, accrual float32, uploadedAt time.Time) *types.Order {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		gm.logger.Error(err)
	}
	return &types.Order{
		Number:     number,
		CustomerID: customerID,
		Accrual:    accrual,
		Status:     status,
		UploadedAt: uploadedAt.In(loc).Format(time.RFC3339),
	}
}

func (gm *GopherMartApp) AppendNewOrder(ctx context.Context, customer *Customer, order *types.Order, accrual float32) error {
	tx, err := gm.Storage.Begin(ctx)
	if err != nil {
		return err
	}
	gm.logger.Infof("store new order %s", order.Number)
	if err = gm.InsertOrder(ctx, order); err != nil {
		gm.logger.Errorf("error inserting order: %s", err.Error())
		if errTx := tx.Rollback(); errTx != nil {
			gm.logger.Errorf("error rolling back transaction: %s", errTx.Error())
		}
		return err
	}
	newBalance, err := gm.UpdateCustomerBalance(ctx, customer, accrual)
	if err != nil {
		gm.logger.Errorf("error update customer balance: %s", err.Error())
		if errTx := tx.Rollback(); errTx != nil {
			gm.logger.Errorf("error rolling back transaction: %s", errTx.Error())
		}
		return err
	}
	gm.logger.Infof("New customer balance: %f", newBalance)
	err = tx.Commit()
	if err != nil {
		gm.logger.Errorf("error committing withdraw: %s", err.Error())
	}
	return err
}

func (gm *GopherMartApp) InsertOrder(ctx context.Context, order *types.Order) error {
	if _, err := gm.Storage.Insert(ctx, sqlInsertOrder, order.Number, order.CustomerID, order.Accrual); err != nil {
		return err
	}
	return nil
}

func (gm *GopherMartApp) GetOrder(ctx context.Context, number string) (*types.Order, error) {
	var order types.Order
	if err := gm.Storage.Get(ctx, &order, sqlGetOrderByID, number); err != nil {
		gm.logger.Warnf("Order possible not found for number: %s", number)
		return nil, err
	}
	return &order, nil
}

func (gm *GopherMartApp) CheckExistsOrder(ctx context.Context, number string, customer *Customer) error {
	var order types.Order
	if err := gm.Storage.Get(ctx, &order, sqlGetOrderByID, number); err != nil {
		fmt.Println(err.Error())
		return nil
	}
	if customer.Id != order.CustomerID {
		return types.ErrOrderAnotherCustomer
	}
	return types.ErrOrderAlreadyExists
}

func (gm *GopherMartApp) GetCustomerOrders(ctx context.Context, customerID int64) ([]types.Order, error) {
	rows, err := gm.Storage.Find(ctx, sqlGetCustomerOrders, customerID)
	defer func() {
		err := rows.Close()
		gm.logger.Error(err.Error())
	}()
	if err != nil {
		return nil, err
	}
	var orders []types.Order
	var number string
	var uploadedAt time.Time
	var status string
	var accrual float32
	for rows.Next() {
		err := rows.Scan(&number, &uploadedAt, &status, &accrual)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		orders = append(orders, *gm.NewOrder(
			number,
			customerID,
			status,
			accrual,
			uploadedAt,
		))
	}
	if err = rows.Err(); err != nil {
		gm.logger.Warn(err.Error())
	}
	gm.logger.Infof("found %d orders; customer %d", len(orders), customerID)
	return orders, nil
}
