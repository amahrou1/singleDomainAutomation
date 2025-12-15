package modules

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"subdomain-recon/config"
	"subdomain-recon/utils"
)

// Feroxbuster module for directory and file bruteforcing
type Feroxbuster struct {
	config   *config.Config
	fxConfig *config.FeroxbusterConfig
}

// NewFeroxbuster creates a new Feroxbuster module
func NewFeroxbuster(cfg *config.Config) *Feroxbuster {
	fxConfig := cfg.NewFeroxbusterConfig()
	return &Feroxbuster{
		config:   cfg,
		fxConfig: fxConfig,
	}
}

// Run executes the feroxbuster module
func (fx *Feroxbuster) Run() error {
	moduleName := "Feroxbuster (Directory/File Bruteforce)"
	utils.LogModuleStart(moduleName)

	startTime := time.Now()

	// Check if feroxbuster is installed
	if err := fx.checkFeroxbuster(); err != nil {
		utils.LogModuleSkip(moduleName, err.Error())
		return err
	}

	// Check if wordlist exists
	if err := fx.checkWordlist(); err != nil {
		utils.LogModuleSkip(moduleName, err.Error())
		return err
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(fx.config.OutputDir, 0755); err != nil {
		utils.LogModuleSkip(moduleName, fmt.Sprintf("Failed to create output directory: %v", err))
		return err
	}

	// Run feroxbuster with timeout
	err := fx.runFeroxbuster()

	duration := time.Since(startTime)
	utils.LogModuleEnd(moduleName, err, duration)

	return err
}

// checkFeroxbuster checks if feroxbuster is installed
func (fx *Feroxbuster) checkFeroxbuster() error {
	_, err := exec.LookPath("feroxbuster")
	if err != nil {
		return fmt.Errorf("feroxbuster is not installed or not in PATH")
	}
	return nil
}

// checkWordlist checks if the wordlist file exists
func (fx *Feroxbuster) checkWordlist() error {
	if _, err := os.Stat(fx.fxConfig.Wordlist); os.IsNotExist(err) {
		return fmt.Errorf("wordlist not found at %s", fx.fxConfig.Wordlist)
	}
	return nil
}

// runFeroxbuster executes the feroxbuster command with timeout
func (fx *Feroxbuster) runFeroxbuster() error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), fx.config.Timeout)
	defer cancel()

	// Prepare status codes as comma-separated string
	statusCodes := strings.Join(fx.fxConfig.StatusCodes, ",")

	// Build the command arguments
	args := []string{
		"-u", fx.config.Target,
		"-w", fx.fxConfig.Wordlist,
		"-s", statusCodes,
		"-t", fmt.Sprintf("%d", fx.fxConfig.Threads),
		"-o", fx.fxConfig.OutputFile,
	}

	// Add SSL skip if enabled
	if fx.fxConfig.SkipSSL {
		args = append(args, "-k")
	}

	// Add no-recursion if depth is 0
	if fx.fxConfig.RecursionDepth == 0 {
		args = append(args, "-n")
	}

	// Add no-state if enabled
	if fx.fxConfig.NoState {
		args = append(args, "--no-state")
	}

	utils.InfoLogger.Printf("Running: feroxbuster %s", strings.Join(args, " "))
	utils.InfoLogger.Printf("Output will be saved to: %s", fx.fxConfig.OutputFile)
	utils.InfoLogger.Printf("Timeout: %s", fx.config.Timeout)

	// Create the command with context
	cmd := exec.CommandContext(ctx, "feroxbuster", args...)

	// Set output to both stdout and file
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()

	// Check if context deadline exceeded (timeout)
	if ctx.Err() == context.DeadlineExceeded {
		utils.WarningLogger.Printf("feroxbuster timed out after %s", fx.config.Timeout)
		// Check if any output was generated
		if fx.checkOutputFile() {
			utils.InfoLogger.Printf("Partial results saved to %s", fx.fxConfig.OutputFile)
			return nil // Don't treat timeout as error if we have partial results
		}
		return fmt.Errorf("feroxbuster timed out after %s with no results", fx.config.Timeout)
	}

	// Check for other errors
	if err != nil {
		// Check if output file was created despite error
		if fx.checkOutputFile() {
			utils.WarningLogger.Printf("feroxbuster encountered an error but produced output: %v", err)
			return nil // Don't fail the module if we have output
		}
		return fmt.Errorf("feroxbuster failed: %v", err)
	}

	// Verify output file exists and has content
	if !fx.checkOutputFile() {
		return fmt.Errorf("feroxbuster completed but no output file was generated")
	}

	// Get file info
	fileInfo, err := os.Stat(fx.fxConfig.OutputFile)
	if err == nil {
		utils.SuccessLogger.Printf("Output saved to: %s (Size: %d bytes)", fx.fxConfig.OutputFile, fileInfo.Size())
	}

	return nil
}

// checkOutputFile checks if the output file exists and has content
func (fx *Feroxbuster) checkOutputFile() bool {
	fileInfo, err := os.Stat(fx.fxConfig.OutputFile)
	if err != nil {
		return false
	}
	return fileInfo.Size() > 0
}

// GetOutputPath returns the absolute path to the output file
func (fx *Feroxbuster) GetOutputPath() (string, error) {
	return filepath.Abs(fx.fxConfig.OutputFile)
}
