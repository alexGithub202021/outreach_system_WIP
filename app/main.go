package main

import (
	"fmt"
	"log/slog"
	lib "myapp/lib"
	"time"
)

var operation = "aggTrade"
var url = fmt.Sprintf("wss://stream.binance.com:9443/ws/btcusdt@%s", operation)
var logPath = "app/logs/logs.json"

func main() {

	cleanup, err := lib.InitGlobalLogger(logPath)
	if nil != err {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer cleanup()

	slog.Info("go app is running")

	testSleep(1800)

}

func testSleep(dur int) {

	duration := time.Duration(dur)

	slog.Info("Starting...")
	time.Sleep(duration * time.Second) // Pause for 5 seconds
	slog.Info("Resuming...")
}
