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

	// Write target header to output file
	ds.writeTargetHeader()

	// Create temp output file for this target
	tempFile := ds.dsConfig.OutputFile + ".temp"

	// Build the command arguments (output to temp file)
	args := []string{
		"-u", ds.config.Target,
		"-e", extensions,
		"-i", statusCodes,
		"-t", fmt.Sprintf("%d", ds.dsConfig.Threads),
		"-o", tempFile,
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

	// Set output to stdout only
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()

	// Append temp file results to main output file
	defer os.Remove(tempFile) // Clean up temp file

	// Check if context deadline exceeded (timeout)
	if ctx.Err() == context.DeadlineExceeded {
		utils.WarningLogger.Printf("dirsearch timed out after %s", ds.config.Timeout)
		// Try to append partial results if they exist
		if ds.appendTempToMain(tempFile) {
			utils.InfoLogger.Printf("Partial results appended to %s", ds.dsConfig.OutputFile)
			return nil
		}
		return fmt.Errorf("dirsearch timed out after %s with no results", ds.config.Timeout)
	}

	// Check for other errors
	if err != nil {
		// Try to append results even if there was an error
		if ds.appendTempToMain(tempFile) {
			utils.WarningLogger.Printf("dirsearch encountered an error but produced output: %v", err)
			return nil // Don't fail the module if we have output
		}
		return fmt.Errorf("dirsearch failed: %v", err)
	}

	// Append temp file to main output file
	if !ds.appendTempToMain(tempFile) {
		return fmt.Errorf("dirsearch completed but no output file was generated")
	}

	// Get file info
	fileInfo, err := os.Stat(ds.dsConfig.OutputFile)
	if err == nil {
		utils.SuccessLogger.Printf("Output appended to: %s (Total size: %d bytes)", ds.dsConfig.OutputFile, fileInfo.Size())
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

// writeTargetHeader writes a separator and target info to the output file
func (ds *Dirsearch) writeTargetHeader() error {
	f, err := os.OpenFile(ds.dsConfig.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	header := fmt.Sprintf("\n\n%s\n[DIRSEARCH] Target: %s\n%s\n",
		strings.Repeat("=", 80),
		ds.config.Target,
		strings.Repeat("=", 80))

	_, err = f.WriteString(header)
	return err
}

// appendTempToMain appends the temporary file content to the main output file
func (ds *Dirsearch) appendTempToMain(tempFile string) bool {
	// Check if temp file exists and has content
	tempInfo, err := os.Stat(tempFile)
	if err != nil || tempInfo.Size() == 0 {
		return false
	}

	// Read temp file
	tempContent, err := os.ReadFile(tempFile)
	if err != nil {
		return false
	}

	// Append to main file
	f, err := os.OpenFile(ds.dsConfig.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Write(tempContent)
	if err != nil {
		return false
	}

	// Add a newline separator after the results
	f.WriteString("\n")

	return true
}
