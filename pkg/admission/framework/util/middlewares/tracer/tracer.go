package tracer

import (
	"context"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	"github.com/caicloud/nirvana/definition"
)

func New(tracer *processor.Tracer) func(context.Context, definition.Chain) error {
	if tracer == nil {
		return func(_ context.Context, _ definition.Chain) error { return nil }
	}
	return func(ctx context.Context, chain definition.Chain) error {
		_, e := tracer.DoWithTracing(func() error {
			return chain.Continue(ctx)
		})
		return e
	}
}
