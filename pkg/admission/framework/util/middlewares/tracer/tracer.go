package tracer

import (
	"sync/atomic"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/errors"
)

type Tracer struct {
	total   ExecTracer
	success ExecTracer
	failed  ExecTracer
}

func (t *Tracer) Update(count uint64, cost time.Duration, err errors.APIStatus) {
	t.total.Update(count, cost)
	if err == nil {
		t.success.Update(count, cost)
	} else {
		t.failed.Update(count, cost)
	}
}

func (t *Tracer) DoWithTracing(f func() errors.APIStatus) (cost time.Duration, err errors.APIStatus) {
	now := time.Now()
	err = f()
	cost = time.Since(now)
	go t.Update(1, cost, err)
	return
}

func (t *Tracer) GetTotalExecTracer() *ExecTracer   { return &t.total }
func (t *Tracer) GetSuccessExecTracer() *ExecTracer { return &t.success }
func (t *Tracer) GetFailedExecTracer() *ExecTracer  { return &t.failed }

type ExecTracer struct {
	count       uint64
	millisecond uint64
}

func (t *ExecTracer) Update(count uint64, cost time.Duration) {
	costMillisecond := uint64(cost / time.Millisecond)
	atomic.AddUint64(&t.count, count)
	atomic.AddUint64(&t.millisecond, costMillisecond)
}

func (t *ExecTracer) GetCount() uint64       { return atomic.LoadUint64(&t.count) }
func (t *ExecTracer) GetMillisecond() uint64 { return atomic.LoadUint64(&t.millisecond) }
func (t *ExecTracer) GetAverageMillisecond() uint64 {
	count := t.GetCount()
	millisecond := t.GetMillisecond()
	if count == 0 {
		return 0
	}
	return millisecond / count
}
