package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var (
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
	ErrorLogger   *log.Logger
	SuccessLogger *log.Logger
)

// InitLogger initializes the loggers
func InitLogger() {
	InfoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	WarningLogger = log.New(os.Stdout, "[WARNING] ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime)
	SuccessLogger = log.New(os.Stdout, "[SUCCESS] ", log.Ldate|log.Ltime)
}

// LogModuleStart logs the start of a module
func LogModuleStart(moduleName string) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	InfoLogger.Printf("Starting Module: %s", moduleName)
	fmt.Println(strings.Repeat("=", 60))
}

// LogModuleEnd logs the end of a module
func LogModuleEnd(moduleName string, err error, duration time.Duration) {
	if err != nil {
		ErrorLogger.Printf("Module '%s' failed: %v (Duration: %s)", moduleName, err, duration)
	} else {
		SuccessLogger.Printf("Module '%s' completed successfully (Duration: %s)", moduleName, duration)
	}
	fmt.Println(strings.Repeat("=", 60) + "\n")
}

// LogModuleSkip logs when a module is skipped
func LogModuleSkip(moduleName string, reason string) {
	WarningLogger.Printf("Module '%s' skipped: %s", moduleName, reason)
	fmt.Println(strings.Repeat("=", 60) + "\n")
}
