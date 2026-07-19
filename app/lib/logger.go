package lib

import (
	"io"
	"log/slog"
	"os"
)

// InitGlobalLogger initializes the global logger and returns a cleanup function.
//
// When logPath is empty (the Lambda / container case) the logger writes to
// stdout only. stdout is captured by CloudWatch, and the Lambda filesystem is
// read-only except /tmp, so file logging must be avoided there.
func InitGlobalLogger(logPath string) (func(), error) {
	var handler slog.Handler
	if logPath == "" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
		cleanup := func() {}
		slog.SetDefault(slog.New(handler))
		return cleanup, nil
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	multiWriter := io.MultiWriter(os.Stdout, file)
	handler = slog.NewJSONHandler(multiWriter, nil)
	slog.SetDefault(slog.New(handler))

	cleanup := func() {
		file.Close()
	}

	return cleanup, nil
}
