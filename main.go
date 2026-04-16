package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"subdomain-recon/config"
	"subdomain-recon/modules"
	"subdomain-recon/utils"
)

const (
	Version    = "2.0"
	ConfigFile = "config.yaml"
)

func main() {
	// Initialize logger
	utils.InitLogger()

	// Define flags
	showHelp := flag.Bool("h", false, "Show help message")
	showVersion := flag.Bool("v", false, "Show version")
	configPath := flag.String("c", ConfigFile, "Path to config file")
	domain := flag.String("d", "", "Single domain to scan")
	listFile := flag.String("l", "", "File containing list of domains (one per line)")
	outputDir := flag.String("o", "", "Output directory for all results (overrides default ./fuzzing-output)")
	timeoutFlag := flag.Duration("t", 0, "Per-module timeout per target (e.g., 10m, 1h). Overrides config.yaml value. 0 = use config")

	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("Single Subdomain Reconnaissance Tool v%s\n", Version)
		os.Exit(0)
	}

	// Show help
	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Print banner
	printBanner()

	// Determine targets
	var targets []string

	if *domain != "" {
		// Single domain mode
		targets = []string{*domain}
	} else if *listFile != "" {
		// List file mode
		var err error
		targets, err = readTargetsFromFile(*listFile)
		if err != nil {
			utils.ErrorLogger.Printf("Failed to read targets from file: %v", err)
			os.Exit(1)
		}
	} else {
		// Check for positional argument (backward compatibility)
		args := flag.Args()
		if len(args) >= 1 {
			targets = []string{args[0]}
		} else {
			fmt.Println("❌ Error: No target specified\n")
			printUsage()
			os.Exit(1)
		}
	}

	// Validate we have targets
	if len(targets) == 0 {
		utils.ErrorLogger.Fatal("No targets found to scan")
	}

	// Print targets summary
	if len(targets) == 1 {
		utils.InfoLogger.Printf("Target: %s", targets[0])
	} else {
		utils.InfoLogger.Printf("Targets: %d subdomains to scan", len(targets))
		fmt.Println("\n[Targets List]")
		fmt.Println("─────────────────────────────────────────────────────────")
		for i, target := range targets {
			fmt.Printf("  %d. %s\n", i+1, target)
		}
		fmt.Println("─────────────────────────────────────────────────────────")
	}

	// Start total timer
	totalStart := time.Now()

	// Process each target sequentially
	successCount := 0
	failCount := 0

	for i, target := range targets {
		if len(targets) > 1 {
			fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
			fmt.Printf("║  Processing Target %d/%d: %-32s║\n", i+1, len(targets), truncateString(target, 32))
			fmt.Println("╚═══════════════════════════════════════════════════════════╝\n")
		}

		// Load configuration for this target
		cfg, err := config.LoadConfig(target, *configPath, *outputDir, *timeoutFlag)
		if err != nil {
			utils.ErrorLogger.Printf("Failed to load configuration for %s: %v", target, err)
			failCount++
			continue
		}

		if i == 0 || len(targets) == 1 {
			utils.SuccessLogger.Printf("Configuration loaded from: %s", *configPath)
		}

		// Print configuration for first target only (to avoid spam)
		if i == 0 {
			printConfig(cfg)
		}

		// Run modules for this target
		targetSuccess := runModules(cfg)

		if targetSuccess {
			successCount++
		} else {
			failCount++
		}

		// Add separator between targets
		if len(targets) > 1 && i < len(targets)-1 {
			fmt.Println("\n" + strings.Repeat("─", 60))
		}
	}

	// Print summary
	totalDuration := time.Since(totalStart)
	printFinalSummary(totalDuration, len(targets), successCount, failCount)
}

// readTargetsFromFile reads targets from a file (one per line)
func readTargetsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var targets []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		targets = append(targets, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no valid targets found in file")
	}

	return targets, nil
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║     Single Subdomain Reconnaissance Tool                 ║
║     Version: ` + Version + `                                          ║
║     Modules: Feroxbuster + Dirsearch + Crawling           ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func printUsage() {
	fmt.Println("Usage: subdomain-recon [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -d <domain>     Scan a single domain")
	fmt.Println("  -l <file>       Scan multiple domains from a file")
	fmt.Println("  -c <file>       Path to config file (default: config.yaml)")
	fmt.Println("  -o <dir>        Output directory for all results (default: ./fuzzing-output)")
	fmt.Println("  -t <duration>   Per-module timeout per target (e.g., 10m, 1h). Overrides config")
	fmt.Println("  -h              Show help message")
	fmt.Println("  -v              Show version")
	fmt.Println("\nExamples:")
	fmt.Println("  # Single domain")
	fmt.Println("  subdomain-recon -d https://api.example.com")
	fmt.Println("")
	fmt.Println("  # Multiple domains from file")
	fmt.Println("  subdomain-recon -l subdomains.txt")
	fmt.Println("")
	fmt.Println("  # With custom output directory")
	fmt.Println("  subdomain-recon -l subdomains.txt -o /root/target/output")
	fmt.Println("")
	fmt.Println("  # 10 minute cap per module per target")
	fmt.Println("  subdomain-recon -l subdomains.txt -o /root/target/output -t 10m")
	fmt.Println("")
	fmt.Println("  # With custom config")
	fmt.Println("  subdomain-recon -d https://api.example.com -c custom.yaml")
	fmt.Println("")
	fmt.Println("  # Backward compatible (positional argument)")
	fmt.Println("  subdomain-recon https://api.example.com")
}

func printHelp() {
	printBanner()
	printUsage()
	fmt.Println("\nDescription:")
	fmt.Println("  A modular reconnaissance tool for scanning single or multiple subdomains.")
	fmt.Println("  Focuses on content discovery and path fuzzing.")
	fmt.Println("\nInput Modes:")
	fmt.Println("  1. Single Domain: Use -d flag to scan one subdomain")
	fmt.Println("     Example: subdomain-recon -d https://api.test.com")
	fmt.Println("")
	fmt.Println("  2. Multiple Domains: Use -l flag with a file containing domains")
	fmt.Println("     Example: subdomain-recon -l targets.txt")
	fmt.Println("")
	fmt.Println("     File format (one domain per line):")
	fmt.Println("       https://api.example.com")
	fmt.Println("       https://admin.example.com")
	fmt.Println("       https://dev.example.com")
	fmt.Println("       # Comments start with #")
	fmt.Println("")
	fmt.Println("  3. Backward Compatible: Provide domain as positional argument")
	fmt.Println("     Example: subdomain-recon https://api.test.com")
	fmt.Println("\nModules:")
	fmt.Println("  1. Feroxbuster - Fast directory/file bruteforcing")
	fmt.Println("  2. Dirsearch   - Advanced path fuzzing with multiple extensions")
	fmt.Println("  3. Crawling    - Web crawling & URL discovery (Katana, Hakrawler, GAU)")
	fmt.Println("\nConfiguration:")
	fmt.Println("  Edit 'config.yaml' to customize:")
	fmt.Println("    - Wordlists")
	fmt.Println("    - Status codes")
	fmt.Println("    - Extensions (dirsearch)")
	fmt.Println("    - Timeout settings")
	fmt.Println("    - Enable/disable modules")
	fmt.Println("    - Output file names")
	fmt.Println("\nOutput:")
	fmt.Println("  Results are saved to ./fuzzing-output/ by default, or to the")
	fmt.Println("  directory specified with the -o flag:")
	fmt.Println("    - feroxbuster.txt     - Feroxbuster results (all targets, single file)")
	fmt.Println("    - dirsearch.txt       - Dirsearch results (all targets, single file)")
	fmt.Println("    - crawling-result.txt - Crawling results (all targets, single file)")
	fmt.Println("\n  Example: subdomain-recon -l subs.txt -o /root/target/output")
	fmt.Println("\nFor more information, see README.md")
}

func printConfig(cfg *config.Config) {
	fmt.Println("\n[Configuration]")
	fmt.Println("─────────────────────────────────────────────────────────")
	fmt.Printf("Target:       %s\n", cfg.Target)
	fmt.Printf("Output Dir:   %s\n", cfg.OutputDir)
	fmt.Printf("Timeout:      %s (per module)\n", cfg.Timeout)
	fmt.Println("\n[Enabled Modules]")
	fmt.Printf("  Feroxbuster:  %v\n", cfg.EnableModules["feroxbuster"])
	fmt.Printf("  Dirsearch:    %v\n", cfg.EnableModules["dirsearch"])
	fmt.Printf("  Crawling:     %v\n", cfg.EnableModules["crawling"])

	if cfg.EnableModules["feroxbuster"] {
		fxCfg := cfg.NewFeroxbusterConfig()
		fmt.Println("\n[Feroxbuster Settings]")
		fmt.Printf("  Wordlist:     %s\n", fxCfg.Wordlist)
		fmt.Printf("  Status Codes: %s\n", fxCfg.StatusCodes)
		fmt.Printf("  Threads:      %d\n", fxCfg.Threads)
		fmt.Printf("  Output:       %s\n", fxCfg.OutputFile)
	}

	if cfg.EnableModules["dirsearch"] {
		dsCfg := cfg.NewDirsearchConfig()
		fmt.Println("\n[Dirsearch Settings]")
		if dsCfg.Wordlist != "" {
			fmt.Printf("  Wordlist:     %s\n", dsCfg.Wordlist)
		} else {
			fmt.Printf("  Wordlist:     (default)\n")
		}
		fmt.Printf("  Extensions:   %d configured\n", len(dsCfg.Extensions))
		fmt.Printf("  Status Codes: %s\n", dsCfg.StatusCodes)
		fmt.Printf("  Threads:      %d\n", dsCfg.Threads)
		fmt.Printf("  Output:       %s\n", dsCfg.OutputFile)
	}

	if cfg.EnableModules["crawling"] {
		crawlCfg := cfg.NewCrawlingConfig()
		fmt.Println("\n[Crawling Settings]")
		fmt.Printf("  Tools:        Katana, Hakrawler, GAU\n")
		fmt.Printf("  Depth:        %d\n", crawlCfg.Katana.Depth)
		fmt.Printf("  Concurrency:  %d\n", crawlCfg.Katana.Concurrency)
		fmt.Printf("  Rate Limit:   %d req/s\n", crawlCfg.Katana.RateLimit)
		fmt.Printf("  Pipeline:     Filter → Unique (anew) → Similar (uro) → Live (httpx)\n")
		fmt.Printf("  Output:       %s\n", crawlCfg.OutputFile)
	}

	fmt.Println("─────────────────────────────────────────────────────────\n")
}

func runModules(cfg *config.Config) bool {
	moduleCount := 0
	successCount := 0
	failCount := 0

	// Module 1: Feroxbuster
	if cfg.EnableModules["feroxbuster"] {
		moduleCount++
		feroxbuster := modules.NewFeroxbuster(cfg)
		err := feroxbuster.Run()
		if err != nil {
			failCount++
			utils.WarningLogger.Printf("Feroxbuster module completed with errors: %v", err)
		} else {
			successCount++
			outputPath, _ := feroxbuster.GetOutputPath()
			utils.SuccessLogger.Printf("Feroxbuster results: %s", outputPath)
		}
	}

	// Module 2: Dirsearch
	if cfg.EnableModules["dirsearch"] {
		moduleCount++
		dirsearch := modules.NewDirsearch(cfg)
		err := dirsearch.Run()
		if err != nil {
			failCount++
			utils.WarningLogger.Printf("Dirsearch module completed with errors: %v", err)
		} else {
			successCount++
			outputPath, _ := dirsearch.GetOutputPath()
			utils.SuccessLogger.Printf("Dirsearch results: %s", outputPath)
		}
	}

	// Module 3: Web Crawling
	if cfg.EnableModules["crawling"] {
		moduleCount++
		crawling := modules.NewCrawling(cfg)
		err := crawling.Run()
		if err != nil {
			failCount++
			utils.WarningLogger.Printf("Crawling module completed with errors: %v", err)
		} else {
			successCount++
			outputPath, _ := crawling.GetOutputPath()
			utils.SuccessLogger.Printf("Crawling results: %s", outputPath)
		}
	}

	// Print module summary for this target
	fmt.Println("\n[Module Execution Summary]")
	fmt.Println("─────────────────────────────────────────────────────────")
	fmt.Printf("Total Modules:      %d\n", moduleCount)
	fmt.Printf("Successful:         %d\n", successCount)
	fmt.Printf("Failed/Skipped:     %d\n", failCount)
	fmt.Println("─────────────────────────────────────────────────────────")

	return failCount == 0
}

func printFinalSummary(duration time.Duration, totalTargets, successCount, failCount int) {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                     Scan Complete                         ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	if totalTargets > 1 {
		fmt.Println("\n[Overall Summary]")
		fmt.Println("─────────────────────────────────────────────────────────")
		fmt.Printf("Total Targets:      %d\n", totalTargets)
		fmt.Printf("Successful:         %d\n", successCount)
		fmt.Printf("Failed:             %d\n", failCount)
		fmt.Println("─────────────────────────────────────────────────────────")
	}

	fmt.Printf("\nTotal Execution Time: %s\n", duration)

	if successCount > 0 {
		fmt.Println("\n✅ Scan complete. Review the output directory for results.")
		fmt.Println("💡 Tip: Review the output files for discovered paths, files, and live URLs")
	}
}
