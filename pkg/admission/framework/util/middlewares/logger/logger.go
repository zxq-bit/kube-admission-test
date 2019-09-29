package logger

import (
	"context"
	"strconv"
	"time"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
)

var (
	green   = "\x1b[1;32m"
	white   = "\x1b[1;37m"
	yellow  = "\x1b[1;33m"
	red     = "\x1b[1;31m"
	blue    = "\x1b[1;34m"
	magenta = "\x1b[1;35m"
	cyan    = "\x1b[1;36m"
	reset   = "\x1b[0m"
)

type Logger interface {
	Infof(string, ...interface{})
}

func New(log Logger) func(context.Context, definition.Chain) error {
	return func(ctx context.Context, chain definition.Chain) error {
		var err error

		start := time.Now()
		req := service.HTTPContextFrom(ctx).Request()
		path := req.URL.Path
		raw := req.URL.RawPath

		err = chain.Continue(ctx)

		w := service.HTTPContextFrom(ctx).ResponseWriter()
		end := time.Now()
		latency := end.Sub(start)
		clientIP := req.RemoteAddr
		method := req.Method
		statusCode := w.StatusCode()
		if raw != "" {
			path = path + "?" + raw
		}

		comment := ""
		if err != nil {
			comment = err.Error()
		}
		log.Infof("%s %13v | %15s | %s %s\n%s",
			httpStatusWithColor(statusCode),
			latency,
			clientIP,
			methodWithColor(method),
			path,
			comment,
		)

		return err
	}
}

func methodWithColor(method string) string {
	return render(colorForMethod(method), method)
}

func httpStatusWithColor(status int) string {
	return render(colorForStatus(status), strconv.Itoa(status))
}

func render(color, msg string) string {
	return color + msg + reset
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS":
		return white
	default:
		return reset
	}
}

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return white
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}
