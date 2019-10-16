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
		_, ke := tracer.DoWithTracing(func() *errors.StatusError {
			e := chain.Continue(ctx)
			if e == nil {
				return nil
			}
			if ke := e.(*errors.StatusError); ke != nil {
				return ke
			}
			return errors.NewInternalError(e)
		})
		if ke != nil {
			return fmt.Errorf("%s", ke.Error())
		}
		return nil
	}
}
