package gm

import (
	"errors"
	"fmt"
	"github.com/GearFramework/gomart2/internal/gm/types"
	"github.com/GearFramework/gomart2/internal/pkg/accrual"
	"strconv"
	"time"
)

func (gm *GopherMartApp) RegisterCustomer(r types.CustomerRegisterRequest) (types.Response, error) {
	gm.logger.Infof("API: register customer")
	if err := validateRegisterRequest(r); err != nil {
		gm.logger.Errorf("invalid form data: %s", err.Error())
		return nil, err
	}
	userID, err := gm.CreateCustomer(r)
	if err != nil {
		gm.logger.Errorf("error created new customer: %s", err.Error())
		return nil, err
	}
	token, err := gm.Auth.CreateToken(userID)
	if err != nil {
		gm.logger.Errorf("error token: %s", err.Error())
		return nil, err
	}
	gm.Auth.SetTokenInCookie(r.GetCtx(), token)
	gm.logger.Infof("customer %s well done registered; new token: %s", r.Login, token)
	return nil, nil
}

func validateRegisterRequest(r types.CustomerRegisterRequest) error {
	if r.Login == "" || r.Password == "" {
		return types.ErrRegisterParamsRequest
	}
	return nil
}

func (gm *GopherMartApp) LoginCustomer(r types.CustomerLoginRequest) (types.Response, error) {
	gm.logger.Infof("API: login customer")
	if err := validateLoginRequest(r); err != nil {
		gm.logger.Errorf("invalid form data: %s", err.Error())
		return nil, err
	}
	customer, err := gm.GetCustomer(r)
	if err != nil {
		gm.logger.Errorf("error login customer %s: %s", r.Login, err.Error())
		return nil, err
	}
	token, err := gm.Auth.CreateToken(customer.ID)
	if err != nil {
		gm.logger.Errorf("error token: %s", err.Error())
		return nil, err
	}
	gm.Auth.SetTokenInCookie(r.GetCtx(), token)
	gm.logger.Infof("customer %s well done login; new token: %s", r.Login, token)
	return nil, nil
}

func validateLoginRequest(r types.CustomerLoginRequest) error {
	if r.Login == "" || r.Password == "" {
		return types.ErrRegisterParamsRequest
	}
	return nil
}

func (gm *GopherMartApp) AddOrder(r types.AddOrderRequest) (types.Response, error) {
	gm.logger.Infof("API: upload customer order %s", r.OrderNumber)
	customerID, err := gm.Auth.AuthCustomer(r.GetCtx())
	if err != nil {
		gm.logger.Errorf("invalid customer: %d", customerID)
		return nil, err
	}
	customer, err := gm.GetCustomerByID(r.GetCtx(), customerID)
	if err != nil {
		gm.logger.Errorf("invalid customer: %d", customerID)
		return nil, err
	}
	gm.logger.Infof("uploading order number %s by customer %s(%d)", r.OrderNumber, customer.Login, customerID)
	if !isValidOrderNumber(r.OrderNumber) {
		gm.logger.Warnf("invalid form data: %s", types.ErrInvalidOrderNumber.Error())
		return nil, types.ErrInvalidOrderNumber
	}
	if err = gm.CheckExistsOrder(r.GetCtx(), r.OrderNumber, customer); err != nil {
		gm.logger.Warnf("upload order %s for customer %s(%d) canceled by: %s",
			r.OrderNumber,
			customer.Login,
			customerID,
			err.Error(),
		)
		return nil, err
	}
	order := gm.NewOrder(r.OrderNumber, customer.ID, accrual.StatusNew, 0, time.Now())
	err = gm.AppendNewOrder(
		r.GetCtx(),
		customer,
		order,
	)

	if errAcc := gm.calcAccrualForOrder(r, customer, order); errAcc != nil {
		gm.logger.Warn(errAcc.Error())
	}

	return nil, err
}

func (gm *GopherMartApp) calcAccrualForOrder(r types.AddOrderRequest, customer *Customer, order *types.Order) error {
	gm.logger.Infof("calculate accrual order %s", r.OrderNumber)
	w, err := gm.Accrual.Calc(r.GetCtx(), r.OrderNumber)
	if err != nil {
		return errors.New(fmt.Sprintf("accrual order %s has rejected by %s", r.OrderNumber, err.Error()))
	}
	gm.logger.Infof("Accrual order status %s and balance %.02f", w.Status, w.Accrual)
	tx, err := gm.Storage.Begin(r.GetCtx())
	if err != nil {
		return errors.New(fmt.Sprintf("error begin transaction: %s", err.Error()))
	}
	if err = gm.UpdateOrderStatusAccrual(r.GetCtx(), order, w.Status, w.Accrual); err != nil {
		if errTx := tx.Rollback(); errTx != nil {
			gm.logger.Errorf("error rolling back transaction: %s", errTx.Error())
		}
		return errors.New(fmt.Sprintf("invalid update status accural for order %s by: %s", r.OrderNumber, err.Error()))
	}
	newBalance, err := gm.UpdateCustomerBalance(r.GetCtx(), customer, w.Accrual)
	if err != nil {
		if errTx := tx.Rollback(); errTx != nil {
			gm.logger.Errorf("error rolling back transaction: %s", errTx.Error())
		}
		return errors.New(fmt.Sprintf("error update customer balance: %s", err.Error()))
	}
	gm.logger.Infof("customer %s(%d) new balance balance: %.02f", customer.Login, customer.ID, newBalance)
	err = tx.Commit()
	if err != nil {
		gm.logger.Errorf("error committing withdraw: %s", err.Error())
	}
	return nil
}

func isValidOrderNumber(num string) bool {
	sum := 0
	nDigits := len(num)
	parity := nDigits % 2
	numb := []byte(num)
	for i := 0; i < nDigits; i++ {
		digit, err := strconv.Atoi(string(numb[i]))
		if err != nil {
			return false
		}
		if (i % 2) == parity {
			digit = digit * 2
		}
		if digit > 9 {
			digit = digit - 9
		}
		sum = sum + digit
	}
	return (sum % 10) == 0
}

func (gm *GopherMartApp) ListOrders(r types.APIRequest) (types.Response, error) {
	gm.logger.Infof("API: list customer orders")
	customerID, err := gm.Auth.AuthCustomer(r.GetCtx())
	if err != nil {
		return nil, err
	}
	customer, err := gm.GetCustomerByID(r.GetCtx(), customerID)
	if err != nil {
		gm.logger.Warnf("error get customer %d: %s", customerID, err.Error())
		return 0, err
	}
	orders, err := gm.GetCustomerOrders(r.GetCtx(), customerID)
	if err != nil {
		gm.logger.Warnf("error get orders for customer %d: %s", customerID, err.Error())
		return nil, err
	}
	gm.logger.Infof("found %d order for customer %s(%d)", len(orders), customer.Login, customerID)
	return orders, nil
}

func (gm *GopherMartApp) GetBalance(r types.APIRequest) (types.Response, error) {
	gm.logger.Infof("API: get customer balance")
	customerID, err := gm.Auth.AuthCustomer(r.GetCtx())
	if err != nil {
		return nil, err
	}
	customer, err := gm.GetCustomerByID(r.GetCtx(), customerID)
	if err != nil {
		gm.logger.Warnf("error get customer %d: %s", customerID, err.Error())
		return 0, err
	}
	gm.logger.Infof("customer %s(%d) balance %.02f", customer.Login, customerID, customer.Balance)
	return &types.CustomerBalanceResponse{Balance: customer.Balance, Withdraw: customer.Withdraw}, nil
}

func (gm *GopherMartApp) ListWithdrawals(r types.APIRequest) (types.Response, error) {
	gm.logger.Infof("API: list customer withdrawls")
	customerID, err := gm.Auth.AuthCustomer(r.GetCtx())
	if err != nil {
		return nil, err
	}
	customer, err := gm.GetCustomerByID(r.GetCtx(), customerID)
	if err != nil {
		gm.logger.Warnf("error get customer %d: %s", customerID, err.Error())
		return 0, err
	}
	orders, err := gm.GetCustomerWithdrawals(r.GetCtx(), customerID)
	if err != nil {
		gm.logger.Warnf("error get withdrawals for customer %d: %s", customerID, err.Error())
		return nil, err
	}
	gm.logger.Infof("found %d withdrawals for customer %s(%d)", len(orders), customer.Login, customerID)
	return orders, nil
}

func (gm *GopherMartApp) Withdraw(r types.CustomerWithdrawRequest) (types.Response, error) {
	gm.logger.Infof("API: withdraw customer; order %s; sum %.02f", r.Order, r.Sum)
	customerID, err := gm.Auth.AuthCustomer(r.GetCtx())
	if err != nil {
		return nil, err
	}
	customer, err := gm.GetCustomerByID(r.GetCtx(), customerID)
	if err != nil {
		gm.logger.Warnf("error get customer %d: %s", customerID, err.Error())
		return 0, err
	}
	if customer.Balance < r.Sum {
		gm.logger.Warnf("error customer %s(%d) points %.02f but requested %.02f",
			customerID,
			customer.Login,
			customer.Withdraw,
			r.Sum,
		)
		return nil, types.ErrNotEnoughPoints
	}
	if !isValidOrderNumber(r.Order) {
		gm.logger.Warnf("invalid requested order %s", r.Order)
		return nil, types.ErrInvalidOrderNumber
	}
	gm.logger.Infof("withdraw for customer %s(%d) by order %s on sum %.02f", customer.Login, customerID, r.Order, r.Sum)
	err = gm.AppendWithdraw(r.GetCtx(), gm.NewWithdraw(r.Order, customerID, r.Sum, time.Now()))
	return nil, err
}
