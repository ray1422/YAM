package utils

import "os"

// GetEnv get env or default value
func GetEnv(key, defaultVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultVal
}
