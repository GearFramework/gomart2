package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/GearFramework/gomart2/internal/pkg/alog"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	timeout                        = time.Duration(3) * time.Second
	StatusNew        StatusAccrual = "NEW"
	StatusRegistered StatusAccrual = "REGISTERED"
	StatusInvalid    StatusAccrual = "INVALID"
	StatusProcessing StatusAccrual = "PROCESSING"
	StatusProcessed  StatusAccrual = "PROCESSED"
)

var (
	ErrInternalError      = errors.New("loyalty system internal error")
	ErrOrderNotRegistered = errors.New("order not registered into loyalty system")
	ErrTooManyRequests    = errors.New("too many requests")
	ErrInvalidContentType = errors.New("invalid content type, required application/json")
	ErrNotProcessed       = errors.New("order not processed")
)

type AccrualClient struct {
	config *Config
	logger *zap.SugaredLogger
}

type ResponseOrders struct {
	Order   string        `json:"order"`
	Status  StatusAccrual `json:"status"`
	Accrual float32       `json:"accrual,omitempty"`
	Timeout time.Duration `json:"-"`
}

type StatusAccrual string

func (sa StatusAccrual) IsValid() bool {
	return sa == StatusProcessing || sa == StatusProcessed
}

func NewClient(addr string) *AccrualClient {
	return &AccrualClient{
		config: &Config{addr: addr},
		logger: alog.NewLogger("info"),
	}
}

func (acc *AccrualClient) Calc(ctx context.Context, number string) (*ResponseOrders, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, acc.config.addr+"/api/orders/"+number, nil)
	if err != nil {
		return nil, err
	}
	client := http.Client{}
	w, err := client.Do(r)
	if err != nil {
		acc.logger.Error(err.Error())
		return nil, err
	}
	err = checkCalcResponse(w)
	defer w.Body.Close()
	if errors.Is(err, ErrTooManyRequests) {
		if tm, errTm := getTimeout(w); err != nil {
			return nil, errTm
		} else {
			return &ResponseOrders{Timeout: tm}, err
		}
	}
	if err != nil {
		acc.logger.Warn("accrual client has error: " + err.Error())
		return nil, err
	}
	data := ResponseOrders{}
	if err := json.NewDecoder(w.Body).Decode(&data); err != nil {
		acc.logger.Error(err.Error())
		return nil, err
	}
	return &data, nil
}

func getTimeout(w *http.Response) (time.Duration, error) {
	h := w.Header.Get("Retry-After")
	tm, err := strconv.Atoi(h)
	if err != nil {
		return 0, ErrInternalError
	}
	return time.Duration(tm) * time.Second, nil
}

func checkCalcResponse(w *http.Response) error {
	if w.StatusCode == http.StatusNoContent {
		return ErrOrderNotRegistered
	}
	if w.StatusCode == http.StatusTooManyRequests {
		return ErrTooManyRequests
	}
	if w.StatusCode != http.StatusOK {
		return ErrInternalError
	}
	if !strings.Contains(w.Header.Get("Content-Type"), "application/json") {
		return ErrInvalidContentType
	}
	return nil
}
