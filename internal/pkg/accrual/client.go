package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/GearFramework/gomart2/internal/pkg/alog"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

const (
	timeout          = time.Duration(3) * time.Second
	StatusNew        = "NEW"
	StatusRegistered = "REGISTERED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"
)

var (
	ErrInternalError      = errors.New("loyalty system internal error")
	ErrOrderNotRegistered = errors.New("order not registered into loyalty system")
	ErrTooManyRequests    = errors.New("too many requests")
	ErrInvalidContentType = errors.New("invalid content type, required application/json")
)

type AccrualClient struct {
	config *Config
	logger *zap.SugaredLogger
}

type ResponseOrders struct {
	Order   string        `json:"order"`
	Status  StatusAccrual `json:"status"`
	Accrual float32       `json:"accrual,omitempty"`
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
		return nil, err
	}
	if err = checkCalcResponse(w); err != nil {
		return nil, err
	}
	defer w.Body.Close()
	data := ResponseOrders{}
	if err := json.NewDecoder(w.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func checkCalcResponse(w *http.Response) error {
	if w.StatusCode == http.StatusNoContent {
		return ErrOrderNotRegistered
	} else if w.StatusCode == http.StatusTooManyRequests {
		return ErrTooManyRequests
	} else if w.StatusCode != http.StatusOK {
		return ErrInternalError
	}
	if !strings.Contains(w.Header.Get("Content-Type"), "application/json") {
		return ErrInvalidContentType
	}
	return nil
}
