package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger = slog.New(slog.DiscardHandler)

func Initialize(level string) error {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		return err
	}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	})

	Log = slog.New(handler)
	return nil
}

func Err(err error) slog.Attr {
	return slog.Any("error", err)
}
