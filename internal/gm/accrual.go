package gm

import (
	"context"
	"errors"
	"fmt"
	"github.com/GearFramework/gomart2/internal/gm/types"
	"github.com/GearFramework/gomart2/internal/pkg/accrual"
)

func (gm *GopherMartApp) calc(ctx context.Context, order any) error {
	w, err := gm.Accrual.Calc(ctx, order.(types.Order).Number)
	if err != nil {
		if errors.Is(err, accrual.ErrTooManyRequests) {
			gm.scheduler.Pause(w.Timeout)
		}
		return err
	}
	tx, err := gm.Storage.Begin(ctx)
	if err != nil {
		msg := fmt.Sprintf("error begin transaction: %s", err.Error())
		return errors.New(msg)
	}
	defer func() {
		if err != nil {
			if errTx := tx.Rollback(); errTx != nil {
				gm.logger.Errorf("error rolling back transaction: %s", errTx.Error())
			}
		}
	}()
	if err = gm.UpdateOrderStatusAccrual(ctx, order.(types.Order), w.Status, w.Accrual); err != nil {
		msg := fmt.Sprintf("invalid update status accural for order %s by: %s",
			order.(types.Order).Number,
			err.Error(),
		)
		return errors.New(msg)
	}
	if w.Status == accrual.StatusInvalid {
		gm.logger.Warnf("order %s was rejected by accrual", order.(types.Order).Number)
		return nil
	}
	if w.Status != accrual.StatusProcessed {
		gm.logger.Warn(accrual.ErrNotProcessed)
		return accrual.ErrNotProcessed
	}
	var newBalance float32
	newBalance, err = gm.UpdateCustomerBalance(ctx, order.(types.Order).CustomerID, w.Accrual)
	if err != nil {
		msg := fmt.Sprintf("error update customer balance: %s", err.Error())
		return errors.New(msg)
	}
	gm.logger.Infof("customerID %d new balance balance: %.02f", order.(types.Order).CustomerID, newBalance)
	err = tx.Commit()
	if err != nil {
		gm.logger.Errorf("error committing withdraw: %s", err.Error())
		return err
	}
	return nil
}
