package main

import (
	"fmt"
	"log/slog"
	logger "myapp/lib"
	"time"
)

var operation = "aggTrade"
var url = fmt.Sprintf("wss://stream.binance.com:9443/ws/btcusdt@%s", operation)

// var logPath = "app/logs/ws_conn_log.json"
var logPath = "app/logs/logs.json"

func main() {

	cleanup, err := logger.InitGlobalLogger(logPath)
	if nil != err {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer cleanup()

	slog.Info("go app is running")

	testSleep(1800)

}

// func getLogger(logPath string) (*slog.Logger, func(), error) {
// 	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	// Pass the file to the JSON handler
// 	logger := slog.New(slog.NewJSONHandler(file, nil))

// 	cleanup := func() {
// 		file.Close()
// 	}

// 	return logger, cleanup, nil
// }

func testSleep(dur int) {

	duration := time.Duration(dur)

	slog.Info("Starting...")
	time.Sleep(duration * time.Second) // Pause for 5 seconds
	slog.Info("Resuming...")
}
