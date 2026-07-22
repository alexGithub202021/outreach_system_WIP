package lib

import (
	"log/slog"
	"strings"

	"github.com/joho/godotenv"
)

func LoadEnv() bool {
	err := godotenv.Load("app/config/.env")
	if err != nil {
		if err.Error() == "no such file or directory" {
			slog.Error("No .env file found. Continuing with defaults.")
			return false
		} else {
			slog.Error("No .env file found. Continuing with defaults.", slog.String("error msg", err.Error()))
			return false
		}
	}
	return true
}

func CapitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
