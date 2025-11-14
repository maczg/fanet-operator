package utils

import (
	"os"
	"strconv"
	"time"
)

func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func ParseDuration(value string) (time.Duration, error) {
	return time.ParseDuration(value)
}

func ParseFloat(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

func ParseInt(value string) (int, error) {
	return strconv.Atoi(value)
}
