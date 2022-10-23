package flags

import (
	"os"
	"strconv"
	"strings"
)

func intFromEnv(s string) int {
	val := os.Getenv(s)
	if val == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func boolFromEnv(s string) bool {
	val := os.Getenv(s)
	return val != "" &&
		val != "0" &&
		strings.ToLower(val) != "false" &&
		strings.ToLower(val) != "no"
}
