package modules

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"subdomain-recon/config"
	"subdomain-recon/utils"
)

// Crawling represents the web crawling module
type Crawling struct {
	config        *config.Config
	crawlConfig   *config.CrawlingConfig
	tempDir       string
	unwantedExts  []string
}

// NewCrawling creates a new crawling module instance
func NewCrawling(cfg *config.Config) *Crawling {
	return &Crawling{
		config:      cfg,
		crawlConfig: cfg.NewCrawlingConfig(),
		tempDir:     cfg.OutputDir + "/.temp-crawl",
		unwantedExts: []string{
			"jpg", "jpeg", "gif", "png", "svg", "ico", "webp", "bmp",
			"css", "woff", "woff2", "ttf", "otf", "eot",
			"mp3", "mp4", "avi", "mov", "wmv", "flv", "wav",
			"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx",
			"zip", "rar", "tar", "gz", "7z",
		},
	}
}

// Run executes the crawling module
func (c *Crawling) Run() error {
	fmt.Println(strings.Repeat("=", 60))
	utils.InfoLogger.Printf("Starting Module: %s", c.GetName())
	fmt.Println(strings.Repeat("=", 60))

	// Check if tools are installed
	if err := c.checkTools(); err != nil {
		utils.WarningLogger.Printf("Module '%s' skipped: %v", c.GetName(), err)
		fmt.Println(strings.Repeat("=", 60))
		return err
	}

	// Create temp directory for intermediate files
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(c.tempDir) // Clean up temp directory

	// Write target header to output file
	c.writeTargetHeader()

	// ONE shared context bounds the entire crawling module (crawlers + pipeline)
	// per target. When it fires, every child command is killed and we fall
	// through to the best-effort partial-result save below.
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	// Run crawling tools (errors are logged but do not abort the module)
	c.runCrawlers(ctx)

	// Process results through filtering pipeline (errors are logged too,
	// so we always get a chance to save whatever stage completed)
	pipelineErr := c.processPipeline(ctx)

	// If we hit the per-target timeout, try to save the best partial result
	if ctx.Err() == context.DeadlineExceeded {
		utils.WarningLogger.Printf("crawling timed out after %s", c.config.Timeout)
		if c.saveBestAvailable() {
			utils.InfoLogger.Printf("Partial results appended to %s", c.crawlConfig.OutputFile)
			return nil
		}
		return fmt.Errorf("crawling timed out after %s with no results", c.config.Timeout)
	}

	// Normal completion: prefer final.txt, fall back to the best intermediate file
	if !c.saveBestAvailable() {
		if pipelineErr != nil {
			return fmt.Errorf("crawling completed but no URLs were collected: %v", pipelineErr)
		}
		return fmt.Errorf("crawling completed but no URLs were collected")
	}

	// Get file info
	fileInfo, err := os.Stat(c.crawlConfig.OutputFile)
	if err == nil {
		utils.SuccessLogger.Printf("Output appended to: %s (Total size: %d bytes)", c.crawlConfig.OutputFile, fileInfo.Size())
	}

	return nil
}

// saveBestAvailable picks the highest-quality pipeline artifact that exists
// on disk and appends it to the main output file.
func (c *Crawling) saveBestAvailable() bool {
	candidates := []string{
		c.tempDir + "/final.txt",    // after httpx (live URLs)
		c.tempDir + "/uro.txt",      // after uro (deduped similar URLs)
		c.tempDir + "/unique.txt",   // after anew (unique URLs)
		c.tempDir + "/filtered.txt", // after extension filter
		c.tempDir + "/combined.txt", // just all crawler outputs concatenated
	}
	for _, f := range candidates {
		if c.appendTempToMain(f) {
			return true
		}
	}
	return false
}

// checkTools verifies that all required tools are installed
func (c *Crawling) checkTools() error {
	tools := []string{"katana", "hakrawler", "gau", "anew", "uro", "httpx"}
	missing := []string{}

	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("required tools not installed: %s", strings.Join(missing, ", "))
	}

	return nil
}

// runCrawlers executes all crawling tools under the shared module context.
func (c *Crawling) runCrawlers(ctx context.Context) error {
	utils.InfoLogger.Println("Running crawlers...")

	// Run Katana scans
	if err := c.runKatanaScans(ctx); err != nil {
		utils.WarningLogger.Printf("Katana failed: %v", err)
	}

	// Bail out early if we've already blown the per-target budget
	if ctx.Err() != nil {
		return nil
	}

	// Run Hakrawler
	if err := c.runHakrawler(ctx); err != nil {
		utils.WarningLogger.Printf("Hakrawler failed: %v", err)
	}

	if ctx.Err() != nil {
		return nil
	}

	// Run GAU
	if err := c.runGAU(ctx); err != nil {
		utils.WarningLogger.Printf("GAU failed: %v", err)
	}

	return nil
}

// runKatanaScans executes multiple Katana scans with different configurations
func (c *Crawling) runKatanaScans(ctx context.Context) error {
	utils.InfoLogger.Println("Running Katana scans (active, passive, parameters, JS)...")

	// Scan 1: Standard crawling with JS and forms
	utils.InfoLogger.Println("  → Katana: Standard crawling with JS and forms")
	if err := c.runKatanaStandard(ctx); err != nil {
		utils.WarningLogger.Printf("Katana standard scan failed: %v", err)
	}

	if ctx.Err() != nil {
		return nil
	}

	// Scan 2: Parameter-focused scan
	utils.InfoLogger.Println("  → Katana: Parameter-focused scan")
	if err := c.runKatanaParams(ctx); err != nil {
		utils.WarningLogger.Printf("Katana parameter scan failed: %v", err)
	}

	if ctx.Err() != nil {
		return nil
	}

	// Scan 3: Passive mode (use archives)
	utils.InfoLogger.Println("  → Katana: Passive mode (archives)")
	if err := c.runKatanaPassive(ctx); err != nil {
		utils.WarningLogger.Printf("Katana passive scan failed: %v", err)
	}

	return nil
}

// runKatanaStandard runs standard Katana crawling
func (c *Crawling) runKatanaStandard(ctx context.Context) error {
	outputFile := c.tempDir + "/katana-standard.txt"

	args := []string{
		"-u", c.config.Target,
		"-d", fmt.Sprintf("%d", c.crawlConfig.Katana.Depth),
		"-jc",                    // JavaScript crawling
		"-kf", "all",             // Crawl all types
		"-aff",                   // Automatic form fill
		"-fx",                    // Form extraction
		"-c", fmt.Sprintf("%d", c.crawlConfig.Katana.Concurrency),
		"-rl", fmt.Sprintf("%d", c.crawlConfig.Katana.RateLimit),
		"-s", "depth-first",
		"-o", outputFile,
		"-silent",
	}

	cmd := exec.CommandContext(ctx, "katana", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// runKatanaParams runs Katana to find URLs with parameters
func (c *Crawling) runKatanaParams(ctx context.Context) error {
	outputFile := c.tempDir + "/katana-params.txt"

	args := []string{
		"-u", c.config.Target,
		"-d", fmt.Sprintf("%d", c.crawlConfig.Katana.Depth),
		"-jc",
		"-f", "qurl", // Only URLs with query parameters
		"-c", fmt.Sprintf("%d", c.crawlConfig.Katana.Concurrency),
		"-rl", fmt.Sprintf("%d", c.crawlConfig.Katana.RateLimit),
		"-o", outputFile,
		"-silent",
	}

	cmd := exec.CommandContext(ctx, "katana", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// runKatanaPassive runs Katana in passive mode
func (c *Crawling) runKatanaPassive(ctx context.Context) error {
	outputFile := c.tempDir + "/katana-passive.txt"

	args := []string{
		"-u", c.config.Target,
		"-passive",              // Passive mode
		"-jc",
		"-o", outputFile,
		"-silent",
	}

	cmd := exec.CommandContext(ctx, "katana", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// runHakrawler executes hakrawler under the shared module context.
func (c *Crawling) runHakrawler(ctx context.Context) error {
	utils.InfoLogger.Println("Running Hakrawler...")

	outputFile := c.tempDir + "/hakrawler.txt"

	// Build command without invalid -plain flag
	args := []string{"-d", "3"}

	utils.InfoLogger.Printf("Running: hakrawler %s", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "hakrawler", args...)

	// Feed target via stdin
	cmd.Stdin = strings.NewReader(c.config.Target + "\n")

	// Create output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	// Run the command
	return cmd.Run()
}

// runGAU executes GAU under the shared module context.
func (c *Crawling) runGAU(ctx context.Context) error {
	utils.InfoLogger.Println("Running GAU (archive URLs)...")

	outputFile := c.tempDir + "/gau.txt"

	// Extract domain from target URL
	domain := c.extractDomain(c.config.Target)

	args := []string{
		"--threads", "50",
		"--blacklist", strings.Join(c.unwantedExts, ","),
		"--o", outputFile,
		domain, // GAU takes domain only, not full URL
	}

	cmd := exec.CommandContext(ctx, "gau", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// processPipeline processes crawled URLs through filtering pipeline
func (c *Crawling) processPipeline(ctx context.Context) error {
	utils.InfoLogger.Println("Processing URLs through filtering pipeline...")

	// Step 1: Combine all crawled URLs (pure Go, no timeout needed)
	combinedFile := c.tempDir + "/combined.txt"
	if err := c.combineFiles(combinedFile); err != nil {
		return fmt.Errorf("failed to combine files: %v", err)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Step 2: Filter unwanted extensions (pure Go)
	filteredFile := c.tempDir + "/filtered.txt"
	if err := c.filterExtensions(combinedFile, filteredFile); err != nil {
		return fmt.Errorf("failed to filter extensions: %v", err)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Step 3: Use anew to get unique URLs
	uniqueFile := c.tempDir + "/unique.txt"
	if err := c.runAnew(ctx, filteredFile, uniqueFile); err != nil {
		return fmt.Errorf("failed to run anew: %v", err)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Step 4: Use uro to remove similar URLs
	uroFile := c.tempDir + "/uro.txt"
	if err := c.runUro(ctx, uniqueFile, uroFile); err != nil {
		return fmt.Errorf("failed to run uro: %v", err)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Step 5: Use httpx to filter live URLs
	finalFile := c.tempDir + "/final.txt"
	if err := c.runHttpx(ctx, uroFile, finalFile); err != nil {
		return fmt.Errorf("failed to run httpx: %v", err)
	}

	utils.SuccessLogger.Println("Filtering pipeline completed")
	return nil
}

// combineFiles combines all crawler output files
func (c *Crawling) combineFiles(outputFile string) error {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	// List of files to combine
	files := []string{
		c.tempDir + "/katana-standard.txt",
		c.tempDir + "/katana-params.txt",
		c.tempDir + "/katana-passive.txt",
		c.tempDir + "/hakrawler.txt",
		c.tempDir + "/gau.txt",
	}

	for _, file := range files {
		// Check if file exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			continue
		}

		// Read and append file contents
		inFile, err := os.Open(file)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(inFile)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				writer.WriteString(line + "\n")
			}
		}
		inFile.Close()
	}

	return nil
}

// filterExtensions removes URLs with unwanted extensions
func (c *Crawling) filterExtensions(inputFile, outputFile string) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url == "" {
			continue
		}

		// Check if URL ends with unwanted extension
		skip := false
		lowerURL := strings.ToLower(url)
		for _, ext := range c.unwantedExts {
			if strings.HasSuffix(lowerURL, "."+ext) {
				skip = true
				break
			}
		}

		if !skip {
			writer.WriteString(url + "\n")
		}
	}

	return scanner.Err()
}

// runAnew uses anew to get unique URLs
func (c *Crawling) runAnew(ctx context.Context, inputFile, outputFile string) error {
	utils.InfoLogger.Println("  → Using anew for unique URLs")

	// anew reads from stdin and writes to specified file
	cmd := exec.CommandContext(ctx, "anew", outputFile)

	inFile, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer inFile.Close()

	cmd.Stdin = inFile
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// runUro uses uro to remove similar URLs
func (c *Crawling) runUro(ctx context.Context, inputFile, outputFile string) error {
	utils.InfoLogger.Println("  → Using uro to remove similar URLs")

	inFile, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	cmd := exec.CommandContext(ctx, "uro")
	cmd.Stdin = inFile
	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// runHttpx uses httpx to filter live URLs
func (c *Crawling) runHttpx(ctx context.Context, inputFile, outputFile string) error {
	utils.InfoLogger.Println("  → Using httpx to filter live URLs")

	args := []string{
		"-l", inputFile,
		"-o", outputFile,
		"-silent",
		"-timeout", "10",
		"-threads", "50",
	}

	cmd := exec.CommandContext(ctx, "httpx", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// extractDomain extracts domain from URL
func (c *Crawling) extractDomain(urlStr string) string {
	// Remove protocol
	domain := strings.TrimPrefix(urlStr, "https://")
	domain = strings.TrimPrefix(domain, "http://")

	// Remove port and path
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	return domain
}

// writeTargetHeader writes a separator and target info to the output file
func (c *Crawling) writeTargetHeader() error {
	f, err := os.OpenFile(c.crawlConfig.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	header := fmt.Sprintf("\n\n%s\n[CRAWLING] Target: %s\n%s\n",
		strings.Repeat("=", 80),
		c.config.Target,
		strings.Repeat("=", 80))

	_, err = f.WriteString(header)
	return err
}

// appendTempToMain appends the final results to the main output file
func (c *Crawling) appendTempToMain(tempFile string) bool {
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
	f, err := os.OpenFile(c.crawlConfig.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

// GetName returns the module name
func (c *Crawling) GetName() string {
	return "Web Crawling & URL Discovery"
}

// GetOutputPath returns the absolute path to the output file
func (c *Crawling) GetOutputPath() (string, error) {
	return filepath.Abs(c.crawlConfig.OutputFile)
}
