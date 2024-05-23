package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/fx"
)

func RunWithGracefulShutdown(app *fx.App) {
	_, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app.Run()
}
