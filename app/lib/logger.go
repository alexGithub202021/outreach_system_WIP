package lib

import (
	"io"
	"log/slog"
	"os"
)

// InitGlobalLogger initializes the global logger and returns a cleanup function
func InitGlobalLogger(logPath string) (func(), error) {
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	multiWriter := io.MultiWriter(os.Stdout, file)
	handler := slog.NewJSONHandler(multiWriter, nil)
	slog.SetDefault(slog.New(handler))

	cleanup := func() {
		file.Close()
	}

	return cleanup, nil
}
