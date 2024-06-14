package gm

import (
	"context"
	"fmt"
	"github.com/GearFramework/gomart2/internal/gm/types"
	"github.com/GearFramework/gomart2/internal/pkg/accrual"
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
	sqlUpdateOrderStatusAccrual = `
		UPDATE gomartspace.orders
		   SET status = $2,
		       accrual = $3
		 WHERE number = $1
	`
)

func (gm *GopherMartApp) NewOrder(number string, customerID int64, status accrual.StatusAccrual, accrual float32, uploadedAt time.Time) *types.Order {
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

func (gm *GopherMartApp) AppendNewOrder(ctx context.Context, customer *Customer, order *types.Order) error {
	gm.logger.Infof("store new order %s", order.Number)
	if err := gm.InsertOrder(ctx, order); err != nil {
		gm.logger.Errorf("error inserting order: %s", err.Error())
		return err
	}
	return nil
}

func (gm *GopherMartApp) InsertOrder(ctx context.Context, order *types.Order) error {
	if _, err := gm.Storage.Insert(ctx, sqlInsertOrder, order.Number, order.CustomerID, order.Accrual); err != nil {
		return err
	}
	return nil
}

func (gm *GopherMartApp) UpdateOrderStatusAccrual(
	ctx context.Context,
	order *types.Order,
	status accrual.StatusAccrual,
	accrual float32,
) error {
	return gm.Storage.Update(ctx, sqlUpdateOrderStatusAccrual, order.Number, status, accrual)
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
	order, err := gm.GetOrder(ctx, number)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	if customer.ID != order.CustomerID {
		return types.ErrOrderAnotherCustomer
	}
	return types.ErrOrderAlreadyExists
}

func (gm *GopherMartApp) GetCustomerOrders(ctx context.Context, customerID int64) ([]types.Order, error) {
	rows, err := gm.Storage.Find(ctx, sqlGetCustomerOrders, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []types.Order
	var number string
	var uploadedAt time.Time
	var status accrual.StatusAccrual
	var balance float32
	for rows.Next() {
		err := rows.Scan(&number, &uploadedAt, &status, &balance)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		orders = append(orders, *gm.NewOrder(
			number,
			customerID,
			status,
			balance,
			uploadedAt,
		))
	}
	if err = rows.Err(); err != nil {
		gm.logger.Warn(err.Error())
	}
	gm.logger.Infof("found %d orders; customer %d", len(orders), customerID)
	return orders, nil
}
