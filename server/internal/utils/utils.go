package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	White   = "\033[97m"
)

func CleanGC() (uint64, uint64) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	before := mem.Alloc / 1_048_576
	runtime.GC() // Cleaning garbage collector
	runtime.ReadMemStats(&mem)
	after := mem.Alloc / 1_048_576
	return before, after
}

func GetExecutableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		panic("impossible to determine the binary path: " + err.Error())
	}
	return filepath.Dir(exePath)
}

const windowSize = 5

func MakePagination(current, totalPages int) []int {
	pages := make([]int, 0, windowSize)

	half := windowSize / 2
	start := current - half
	end := current + half

	if start < 0 {
		end += -start
		start = 0
	}

	if end > totalPages-1 {
		start -= (end - (totalPages - 1))
		end = totalPages - 1
	}
	if start < 0 {
		start = 0
	}

	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	return pages
}
