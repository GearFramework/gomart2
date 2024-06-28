package query

import (
	"context"
	"errors"
	"github.com/GearFramework/gomart2/internal/pkg/accrual"
	"github.com/GearFramework/gomart2/internal/pkg/alog"
	"go.uber.org/zap"
	"sync"
	"time"
)

var (
	ErrEmptyQuery = errors.New("query is empty")
)

type JobCallbackContext func(context.Context, any) error

type Query struct {
	mtx     sync.RWMutex
	client  *accrual.AccrualClient
	packets []any
	logger  *zap.SugaredLogger
	job     JobCallbackContext
	paused  time.Duration
}

func NewQuery(client *accrual.AccrualClient, job JobCallbackContext) *Query {
	return &Query{
		client:  client,
		packets: make([]any, 0),
		logger:  alog.NewLogger("info"),
		job:     job,
	}
}

func (q *Query) Run() {
	for {
		if q.paused > 0 {
			time.Sleep(q.paused)
			q.Continue()
		}
		item, err := q.Pop()
		if errors.Is(err, ErrEmptyQuery) {
			continue
		}
		if err = q.runJob(item); err != nil {
			q.Push(item)
			continue
		}
	}
}

func (q *Query) Pause(tm time.Duration) {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	q.paused = tm
}

func (q *Query) Continue() {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	q.paused = 0
}

func (q *Query) runJob(item any) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return q.job(ctx, item)
}

func (q *Query) Push(item any) {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	q.packets = append(q.packets, item)
}

func (q *Query) Pop() (any, error) {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	if q.IsEmpty() {
		return nil, ErrEmptyQuery
	}
	order := q.packets[0]
	q.packets = q.packets[1:]
	return order, nil
}

func (q *Query) IsEmpty() bool {
	return len(q.packets) == 0
}
