package application

import (
	"context"
	"io/ioutil"
	"strings"

	"qubert/internal/logger"
)

const (
	contextKeyCancelContext = "cancelFunc"
)

func makeContext(ctx context.Context, logger *logger.Logger) context.Context {
	logger.Info("Create context...")

	ctx, cancel := context.WithCancel(ctx)

	return context.WithValue(ctx, contextKeyCancelContext, func() {
		logger.Info("Canceling context...")
		cancel()
	})
}

func cancelContext(ctx context.Context) {
	if f, ok := ctx.Value(contextKeyCancelContext).(func()); ok {
		f()
	}
}

func getHostName() string {
	data, err := ioutil.ReadFile("/etc/hostname")
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(data), "\n")

	if len(lines) > 0 {
		return lines[0]
	}

	return ""
}
