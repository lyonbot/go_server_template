package misc

import "os"

func DefaultString(s1, s2 string) string {
	if s1 != "" {
		return s1
	}
	return s2
}

func Getenv(key, defaults string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaults
}
