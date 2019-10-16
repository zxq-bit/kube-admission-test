package middleware

import (
	"context"
	"time"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
)

type key int

const (
	tenantHeader = "X-Tenant"

	contextKeyTenantID key = iota
)

// Timeout returns a Nirvana middleware that set a timeout of the given duration to the context.
// The function only takes effect if the given timeout is shorter than any existing timeout.
func Timeout(timeout time.Duration) definition.Middleware {
	return func(ctx context.Context, chain definition.Chain) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return chain.Continue(ctx)
	}
}

// TenantID returns a Nirvana middleware that add the tenant ID to the context. The added ID can
// then be retrieved by calling GetTenantID on the context object.
func TenantID(defaultTenant string) definition.Middleware {
	return func(ctx context.Context, chain definition.Chain) error {
		tenantID := service.HTTPContextFrom(ctx).Request().Header.Get(tenantHeader)
		if len(tenantID) == 0 {
			tenantID = defaultTenant
		}
		ctx = context.WithValue(ctx, contextKeyTenantID, tenantID)
		return chain.Continue(ctx)
	}
}

// GetTenantID retrieves the tenant ID from the context, or an empty string if it does not exist.
// The tenant ID should have been stored in the context by the TenantID middleware.
func GetTenantID(ctx context.Context) string {
	// the second returned value indicating if the assertion succeeded
	// if not, the first returned value will be the zero value of type T, which is string here
	ret, _ := ctx.Value(contextKeyTenantID).(string)
	return ret
}
