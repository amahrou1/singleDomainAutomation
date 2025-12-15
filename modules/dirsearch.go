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

// Dirsearch module for advanced web path fuzzing
type Dirsearch struct {
	config   *config.Config
	dsConfig *config.DirsearchConfig
}

// NewDirsearch creates a new Dirsearch module
func NewDirsearch(cfg *config.Config) *Dirsearch {
	dsConfig := cfg.NewDirsearchConfig()
	return &Dirsearch{
		config:   cfg,
		dsConfig: dsConfig,
	}
}

// Run executes the dirsearch module
func (ds *Dirsearch) Run() error {
	moduleName := "Dirsearch (Advanced Path Fuzzing)"
	utils.LogModuleStart(moduleName)

	startTime := time.Now()

	// Check if dirsearch is installed
	if err := ds.checkDirsearch(); err != nil {
		utils.LogModuleSkip(moduleName, err.Error())
		return err
	}

	// Check wordlist if specified
	if ds.dsConfig.Wordlist != "" {
		if err := ds.checkWordlist(); err != nil {
			utils.LogModuleSkip(moduleName, err.Error())
			return err
		}
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(ds.config.OutputDir, 0755); err != nil {
		utils.LogModuleSkip(moduleName, fmt.Sprintf("Failed to create output directory: %v", err))
		return err
	}

	// Run dirsearch with timeout
	err := ds.runDirsearch()

	duration := time.Since(startTime)
	utils.LogModuleEnd(moduleName, err, duration)

	return err
}

// checkDirsearch checks if dirsearch is installed
func (ds *Dirsearch) checkDirsearch() error {
	_, err := exec.LookPath("dirsearch")
	if err != nil {
		return fmt.Errorf("dirsearch is not installed or not in PATH")
	}
	return nil
}

// checkWordlist checks if the wordlist file exists (if specified)
func (ds *Dirsearch) checkWordlist() error {
	if ds.dsConfig.Wordlist == "" {
		return nil // No wordlist specified, dirsearch will use default
	}
	if _, err := os.Stat(ds.dsConfig.Wordlist); os.IsNotExist(err) {
		return fmt.Errorf("wordlist not found at %s", ds.dsConfig.Wordlist)
	}
	return nil
}

// runDirsearch executes the dirsearch command with timeout
func (ds *Dirsearch) runDirsearch() error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ds.config.Timeout)
	defer cancel()

	// Prepare extensions as comma-separated string
	extensions := strings.Join(ds.dsConfig.Extensions, ",")

	// Prepare status codes as comma-separated string
	statusCodes := strings.Join(ds.dsConfig.StatusCodes, ",")

	// Build the command arguments
	// dirsearch -u <url> -e <extensions> --full-url -i <status_codes> -r -o <output>
	args := []string{
		"-u", ds.config.Target,
		"-e", extensions,
		"-i", statusCodes,
		"-t", fmt.Sprintf("%d", ds.dsConfig.Threads),
		"-o", ds.dsConfig.OutputFile,
	}

	// Add wordlist if specified
	if ds.dsConfig.Wordlist != "" {
		args = append(args, "-w", ds.dsConfig.Wordlist)
	}

	// Add full URL option
	if ds.dsConfig.FullURL {
		args = append(args, "--full-url")
	}

	// Add recursive option
	if ds.dsConfig.Recursive {
		args = append(args, "-r")
	}

	utils.InfoLogger.Printf("Running: dirsearch %s", strings.Join(args, " "))
	utils.InfoLogger.Printf("Output will be saved to: %s", ds.dsConfig.OutputFile)
	utils.InfoLogger.Printf("Timeout: %s", ds.config.Timeout)
	utils.InfoLogger.Printf("Extensions: %d configured", len(ds.dsConfig.Extensions))

	// Create the command with context
	cmd := exec.CommandContext(ctx, "dirsearch", args...)

	// Set output to both stdout and file
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()

	// Check if context deadline exceeded (timeout)
	if ctx.Err() == context.DeadlineExceeded {
		utils.WarningLogger.Printf("dirsearch timed out after %s", ds.config.Timeout)
		// Check if any output was generated
		if ds.checkOutputFile() {
			utils.InfoLogger.Printf("Partial results saved to %s", ds.dsConfig.OutputFile)
			return nil // Don't treat timeout as error if we have partial results
		}
		return fmt.Errorf("dirsearch timed out after %s with no results", ds.config.Timeout)
	}

	// Check for other errors
	if err != nil {
		// Check if output file was created despite error
		if ds.checkOutputFile() {
			utils.WarningLogger.Printf("dirsearch encountered an error but produced output: %v", err)
			return nil // Don't fail the module if we have output
		}
		return fmt.Errorf("dirsearch failed: %v", err)
	}

	// Verify output file exists and has content
	if !ds.checkOutputFile() {
		return fmt.Errorf("dirsearch completed but no output file was generated")
	}

	// Get file info
	fileInfo, err := os.Stat(ds.dsConfig.OutputFile)
	if err == nil {
		utils.SuccessLogger.Printf("Output saved to: %s (Size: %d bytes)", ds.dsConfig.OutputFile, fileInfo.Size())
	}

	return nil
}

// checkOutputFile checks if the output file exists and has content
func (ds *Dirsearch) checkOutputFile() bool {
	fileInfo, err := os.Stat(ds.dsConfig.OutputFile)
	if err != nil {
		return false
	}
	return fileInfo.Size() > 0
}

// GetOutputPath returns the absolute path to the output file
func (ds *Dirsearch) GetOutputPath() (string, error) {
	return filepath.Abs(ds.dsConfig.OutputFile)
}
