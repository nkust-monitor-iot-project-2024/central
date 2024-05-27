// Package slogext provides additional slog.Attr constructors.
package slogext

import (
	"log/slog"
)

// Error returns a slog.Attr with the key "error" and the error message.
func Error(err error) slog.Attr {
	return slog.String("error", err.Error())
}
