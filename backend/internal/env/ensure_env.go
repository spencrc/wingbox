package env

import (
	"fmt"
	"os"
)

func Ensureenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Environment variable %s failed to load", key))
	}
	return val
}