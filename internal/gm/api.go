package gm

import (
	"github.com/GearFramework/gomart2/internal/gm/types"
	"strconv"
	"time"
)

func (gm *GopherMartApp) RegisterCustomer(r types.CustomerRegisterRequest) (types.Response, error) {
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
	if err := validateLoginRequest(r); err != nil {
		gm.logger.Errorf("invalid form data: %s", err.Error())
		return nil, err
	}
	customer, err := gm.GetCustomer(r)
	if err != nil {
		gm.logger.Errorf("error login customer: %s", err.Error())
		return nil, err
	}
	token, err := gm.Auth.CreateToken(customer.Id)
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
	customerID, err := gm.Auth.AuthCustomer(r.GetCtx())
	if err != nil {
		return nil, err
	}
	customer, err := gm.GetCustomerByID(r.GetCtx(), customerID)
	if err != nil {
		gm.logger.Errorf("invalid customer: %d", customerID)
		return nil, err
	}
	if !isValidOrderNumber(r.OrderNumber) {
		gm.logger.Errorf("invalid form data: %s", types.ErrInvalidOrderNumber.Error())
		return nil, types.ErrInvalidOrderNumber
	}
	_, err = gm.GetOrder(r.GetCtx(), r.OrderNumber)
	if err == nil {
		gm.logger.Errorf("order %s already exists", r.OrderNumber)
		return nil, types.ErrOrderAlreadyExists
	}
	gm.logger.Infof("Calculate accrual order %s", r.OrderNumber)
	w, err := gm.Accrual.Calc(r.GetCtx(), r.OrderNumber)
	if err != nil {
		gm.logger.Errorf(err.Error())
		return nil, err
	}
	gm.logger.Infof("Accrual order status %s", w.Status)
	var accrual float32
	if w.Status.IsValid() {
		accrual = w.Accrual
	}
	err = gm.AppendNewOrder(
		r.GetCtx(),
		customer,
		gm.NewOrder(r.OrderNumber, customer.Id, string(w.Status), accrual, time.Now()),
		accrual,
	)
	return nil, err
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
	customerID, err := gm.Auth.AuthCustomer(r.GetCtx())
	if err != nil {
		return nil, err
	}
	orders, err := gm.GetCustomerOrders(r.GetCtx(), customerID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (gm *GopherMartApp) GetBalance(r types.APIRequest) (types.Response, error) {
	customerID, err := gm.Auth.AuthCustomer(r.GetCtx())
	if err != nil {
		return nil, err
	}
	customer, err := gm.GetCustomerByID(r.GetCtx(), customerID)
	if err != nil {
		return 0, err
	}
	return &types.CustomerBalanceResponse{Balance: customer.Balance, Withdraw: customer.Withdraw}, nil
}

func (gm *GopherMartApp) ListWithdrawals(r types.APIRequest) (types.Response, error) {
	customerID, err := gm.Auth.AuthCustomer(r.GetCtx())
	if err != nil {
		return nil, err
	}
	orders, err := gm.GetCustomerWithdrawals(r.GetCtx(), customerID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (gm *GopherMartApp) Withdraw(r types.CustomerWithdrawRequest) (types.Response, error) {
	customerID, err := gm.Auth.AuthCustomer(r.GetCtx())
	if err != nil {
		return nil, err
	}
	customer, err := gm.GetCustomerByID(r.GetCtx(), customerID)
	if err != nil {
		return 0, err
	}
	if customer.Balance < r.Sum {
		gm.logger.Errorf("error customer points %.02f but requested %.02f", customer.Withdraw, r.Sum)
		return nil, types.ErrNotEnoughPoints
	}
	if !isValidOrderNumber(r.Order) {
		gm.logger.Errorf("invalid requested order %s", r.Order)
		return nil, types.ErrInvalidOrderNumber
	}
	err = gm.AppendWithdraw(r.GetCtx(), gm.NewWithdraw(r.Order, customerID, r.Sum, time.Now()))
	return nil, err
}
