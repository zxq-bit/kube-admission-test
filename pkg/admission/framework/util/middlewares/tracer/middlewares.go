package tracer

import (
	"context"
	"fmt"

	"github.com/caicloud/nirvana/definition"

	"k8s.io/apimachinery/pkg/api/errors"
)

func New(tracer *Tracer) func(context.Context, definition.Chain) error {
	if tracer == nil {
		return func(_ context.Context, _ definition.Chain) error { return nil }
	}
	return func(ctx context.Context, chain definition.Chain) error {
		_, ke := tracer.DoWithTracing(func() errors.APIStatus {
			e := chain.Continue(ctx)
			if e == nil {
				return nil
			}
			if ke := e.(errors.APIStatus); ke != nil {
				return ke
			}
			return errors.NewInternalError(e)
		})
		return fmt.Errorf("%v", ke)
	}
}
