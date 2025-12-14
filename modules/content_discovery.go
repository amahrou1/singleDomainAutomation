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

// ContentDiscovery module for directory and file bruteforcing
type ContentDiscovery struct {
	config *config.Config
	cdConfig *config.ContentDiscoveryConfig
}

// NewContentDiscovery creates a new ContentDiscovery module
func NewContentDiscovery(cfg *config.Config) *ContentDiscovery {
	cdConfig := config.NewContentDiscoveryConfig(cfg.OutputDir)
	return &ContentDiscovery{
		config:   cfg,
		cdConfig: cdConfig,
	}
}

// Run executes the content discovery module
func (cd *ContentDiscovery) Run() error {
	moduleName := "Content Discovery (feroxbuster)"
	utils.LogModuleStart(moduleName)

	startTime := time.Now()

	// Check if feroxbuster is installed
	if err := cd.checkFeroxbuster(); err != nil {
		utils.LogModuleSkip(moduleName, err.Error())
		return err
	}

	// Check if wordlist exists
	if err := cd.checkWordlist(); err != nil {
		utils.LogModuleSkip(moduleName, err.Error())
		return err
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(cd.config.OutputDir, 0755); err != nil {
		utils.LogModuleSkip(moduleName, fmt.Sprintf("Failed to create output directory: %v", err))
		return err
	}

	// Run feroxbuster with timeout
	err := cd.runFeroxbuster()

	duration := time.Since(startTime)
	utils.LogModuleEnd(moduleName, err, duration)

	return err
}

// checkFeroxbuster checks if feroxbuster is installed
func (cd *ContentDiscovery) checkFeroxbuster() error {
	_, err := exec.LookPath("feroxbuster")
	if err != nil {
		return fmt.Errorf("feroxbuster is not installed or not in PATH")
	}
	return nil
}

// checkWordlist checks if the wordlist file exists
func (cd *ContentDiscovery) checkWordlist() error {
	if _, err := os.Stat(cd.config.WordlistPath); os.IsNotExist(err) {
		return fmt.Errorf("wordlist not found at %s", cd.config.WordlistPath)
	}
	return nil
}

// runFeroxbuster executes the feroxbuster command with timeout
func (cd *ContentDiscovery) runFeroxbuster() error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cd.config.Timeout)
	defer cancel()

	// Prepare status codes as comma-separated string
	statusCodes := strings.Join(cd.cdConfig.StatusCodes, ",")

	// Build the command
	// feroxbuster -u https://subdomain.test.com -w /root/myLists/all.txt -s 200,403,301,302,307 -k -n --no-state
	args := []string{
		"-u", cd.config.Target,
		"-w", cd.config.WordlistPath,
		"-s", statusCodes,
		"-k",           // Skip SSL verification
		"-n",           // Don't scan recursively
		"--no-state",   // Don't save or load state
		"-o", cd.cdConfig.OutputFile, // Output file
	}

	utils.InfoLogger.Printf("Running: feroxbuster %s", strings.Join(args, " "))
	utils.InfoLogger.Printf("Output will be saved to: %s", cd.cdConfig.OutputFile)
	utils.InfoLogger.Printf("Timeout: %s", cd.config.Timeout)

	// Create the command with context
	cmd := exec.CommandContext(ctx, "feroxbuster", args...)

	// Set output to both stdout and file
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()

	// Check if context deadline exceeded (timeout)
	if ctx.Err() == context.DeadlineExceeded {
		utils.WarningLogger.Printf("feroxbuster timed out after %s", cd.config.Timeout)
		// Check if any output was generated
		if cd.checkOutputFile() {
			utils.InfoLogger.Printf("Partial results saved to %s", cd.cdConfig.OutputFile)
			return nil // Don't treat timeout as error if we have partial results
		}
		return fmt.Errorf("feroxbuster timed out after %s with no results", cd.config.Timeout)
	}

	// Check for other errors
	if err != nil {
		// Check if output file was created despite error
		if cd.checkOutputFile() {
			utils.WarningLogger.Printf("feroxbuster encountered an error but produced output: %v", err)
			return nil // Don't fail the module if we have output
		}
		return fmt.Errorf("feroxbuster failed: %v", err)
	}

	// Verify output file exists and has content
	if !cd.checkOutputFile() {
		return fmt.Errorf("feroxbuster completed but no output file was generated")
	}

	// Get file info
	fileInfo, err := os.Stat(cd.cdConfig.OutputFile)
	if err == nil {
		utils.SuccessLogger.Printf("Output saved to: %s (Size: %d bytes)", cd.cdConfig.OutputFile, fileInfo.Size())
	}

	return nil
}

// checkOutputFile checks if the output file exists and has content
func (cd *ContentDiscovery) checkOutputFile() bool {
	fileInfo, err := os.Stat(cd.cdConfig.OutputFile)
	if err != nil {
		return false
	}
	return fileInfo.Size() > 0
}

// GetOutputPath returns the absolute path to the output file
func (cd *ContentDiscovery) GetOutputPath() (string, error) {
	return filepath.Abs(cd.cdConfig.OutputFile)
}
