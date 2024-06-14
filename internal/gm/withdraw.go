package gm

import (
	"context"
	"fmt"
	"github.com/GearFramework/gomart2/internal/gm/types"
	"time"
)

const (
	sqlInsertWithdraw = `
		INSERT INTO gomartspace.withdrawals 
		       (number, customer_id, sum) 
		VALUES ($1, $2, $3)
	`
	sqlGetCustomerWithdrawals = `
		SELECT number,
			   sum,
			   processed_at
		  FROM gomartspace.withdrawals
		 WHERE customer_id = $1
		 ORDER BY processed_at DESC
	`
	sqlUpdateCustomerWithdraw = `
		UPDATE gomartspace.customers
		   SET balance = balance - $1,
		       withdraw = withdraw + $1
		 WHERE id = $2
	`
)

func (gm *GopherMartApp) NewWithdraw(number string, customerID int64, sum float32, processedAt time.Time) *types.Withdraw {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		gm.logger.Error(err)
	}
	return &types.Withdraw{
		Number:      number,
		CustomerID:  customerID,
		Sum:         sum,
		ProcessedAt: processedAt.In(loc).Format(time.RFC3339),
	}
}

func (gm *GopherMartApp) AppendWithdraw(ctx context.Context, withdraw *types.Withdraw) error {
	tx, err := gm.Storage.Begin(ctx)
	if err != nil {
		return err
	}
	if err = gm.InsertWithdraw(ctx, withdraw); err != nil {
		gm.logger.Errorf("inserted withdraw with error: %s", err.Error())
		if errTx := tx.Rollback(); errTx != nil {
			gm.logger.Errorf("error rolling back transaction: %s", errTx.Error())
		}
		return types.ErrOrderAlreadyExists
	}
	if err := gm.UpdateCustomerWithdraw(ctx, withdraw); err != nil {
		gm.logger.Errorf("error updating withdraw: %s", err.Error())
		if errTx := tx.Rollback(); errTx != nil {
			gm.logger.Errorf("error rolling back transaction: %s", errTx.Error())
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		gm.logger.Errorf("error committing withdraw: %s", err.Error())
	}
	return err
}

func (gm *GopherMartApp) InsertWithdraw(ctx context.Context, withdraw *types.Withdraw) error {
	if _, err := gm.Storage.Insert(ctx, sqlInsertWithdraw, withdraw.Number, withdraw.CustomerID, withdraw.Sum); err != nil {
		return err
	}
	return nil
}

func (gm *GopherMartApp) UpdateCustomerWithdraw(ctx context.Context, withdraw *types.Withdraw) error {
	return gm.Storage.Update(ctx, sqlUpdateCustomerWithdraw, withdraw.Sum, withdraw.CustomerID)
}

func (gm *GopherMartApp) GetCustomerWithdrawals(ctx context.Context, customerID int64) ([]types.Withdraw, error) {
	rows, err := gm.Storage.Find(ctx, sqlGetCustomerWithdrawals, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var withdrawals []types.Withdraw
	var number string
	var sum float32
	var processedAt time.Time
	for rows.Next() {
		err := rows.Scan(&number, &sum, &processedAt)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		withdrawals = append(withdrawals, *gm.NewWithdraw(
			number,
			customerID,
			sum,
			processedAt,
		))
	}
	if err = rows.Err(); err != nil {
		gm.logger.Warn(err.Error())
	}
	gm.logger.Infof("found %d withdrawals; customer %d", len(withdrawals), customerID)
	return withdrawals, nil
}
