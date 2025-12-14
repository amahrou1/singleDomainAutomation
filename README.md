# Single Subdomain Reconnaissance Tool

A modular Go-based reconnaissance tool focused on scanning a single subdomain for bug hunting and security testing.

## Current Features

### Module 1: Content Discovery (feroxbuster)
- Directory and file bruteforcing
- Customizable status code filtering (200, 403, 301, 302, 307)
- 5-hour timeout protection
- Robust error handling
- Automatic output saving

## Project Structure

```
singleDomainAutomation/
├── main.go                          # Main entry point
├── config/
│   └── config.go                    # Configuration management
├── modules/
│   └── content_discovery.go         # Content Discovery module
├── utils/
│   └── logger.go                    # Logging utilities
├── output/                          # Output directory for results
│   └── dirs.txt                     # Feroxbuster output
└── subdomain-recon                  # Compiled binary
```

## Prerequisites

1. **Go** (version 1.19 or higher)
   ```bash
   go version
   ```

2. **feroxbuster** - Fast content discovery tool
   ```bash
   # Install on Linux
   curl -sL https://raw.githubusercontent.com/epi052/feroxbuster/main/install-nix.sh | bash

   # Or download from releases
   wget https://github.com/epi052/feroxbuster/releases/latest/download/x86_64-linux-feroxbuster.tar.gz
   tar -xzf x86_64-linux-feroxbuster.tar.gz
   sudo mv feroxbuster /usr/local/bin/
   ```

3. **Wordlist** - Place your wordlist at `/root/myLists/all.txt` or modify the path in config

## Installation

1. Clone or download this repository
2. Build the application:
   ```bash
   go build -o subdomain-recon main.go
   ```

## Usage

### Basic Usage
```bash
./subdomain-recon https://subdomain.test.com
```

### Alternative (without building)
```bash
go run main.go https://subdomain.test.com
```

## How It Works

### Content Discovery Module

The Content Discovery module uses feroxbuster to brute-force directories and files on the target subdomain.

**Command executed:**
```bash
feroxbuster -u https://subdomain.test.com \
  -w /root/myLists/all.txt \
  -s 200,403,301,302,307 \
  -k \
  -n \
  --no-state \
  -o ./output/dirs.txt
```

**Parameters:**
- `-u` : Target URL
- `-w` : Wordlist path
- `-s` : Status codes to include (200, 403, 301, 302, 307)
- `-k` : Skip SSL certificate verification
- `-n` : Don't scan recursively
- `--no-state` : Don't save or load state
- `-o` : Output file path

**Features:**
- ✅ 5-hour timeout protection
- ✅ Error handling (skips module if feroxbuster not found)
- ✅ Checks wordlist existence before running
- ✅ Creates output directory automatically
- ✅ Saves results to `./output/dirs.txt`
- ✅ Continues execution even on partial failures

## Configuration

You can customize settings by editing `config/config.go`:

```go
cfg := config.NewConfig(target)
cfg.OutputDir = "./custom-output"      // Change output directory
cfg.WordlistPath = "/path/to/wordlist" // Change wordlist path
cfg.Timeout = 3 * time.Hour            // Change timeout (default: 5 hours)
```

For Content Discovery specific settings, edit `config.NewContentDiscoveryConfig()`:

```go
StatusCodes: []string{"200", "403", "301", "302", "307", "401", "500"}
```

## Output

Results are saved in the `./output/` directory:
- `dirs.txt` - Feroxbuster output with discovered directories and files

## Error Handling

The tool includes robust error handling:

1. **Missing feroxbuster**: Module skips with a warning
2. **Missing wordlist**: Module skips with error message
3. **Timeout**: Saves partial results if available
4. **Runtime errors**: Logs error but continues execution

## Example Output

```
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║     Single Subdomain Reconnaissance Tool                 ║
║     Version: 1.0                                          ║
║     Module: Content Discovery (feroxbuster)               ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝

[INFO] 2025/12/14 11:49:57 Target: https://subdomain.test.com

[Configuration]
─────────────────────────────────────────────────────────
Target:       https://subdomain.test.com
Output Dir:   ./output
Wordlist:     /root/myLists/all.txt
Timeout:      5h0m0s
─────────────────────────────────────────────────────────

============================================================
[INFO] Starting Module: Content Discovery (feroxbuster)
============================================================
[INFO] Running: feroxbuster -u https://subdomain.test.com -w /root/myLists/all.txt -s 200,403,301,302,307 -k -n --no-state -o ./output/dirs.txt
[INFO] Output will be saved to: ./output/dirs.txt
[INFO] Timeout: 5h0m0s

[feroxbuster output appears here...]

[SUCCESS] Module 'Content Discovery (feroxbuster)' completed successfully
[SUCCESS] Output saved to: ./output/dirs.txt (Size: 15234 bytes)
============================================================

╔═══════════════════════════════════════════════════════════╗
║                     Scan Summary                          ║
╚═══════════════════════════════════════════════════════════╝
Total Execution Time: 15m23s

Recon completed! Check the output directory for results.
```

## Future Modules

This is the first module. Additional modules will be added:
- Module 2: Port Scanning
- Module 3: Technology Detection
- Module 4: SSL/TLS Analysis
- Module 5: JavaScript Analysis
- And more...

## Development

### Adding New Modules

1. Create a new file in `modules/` (e.g., `port_scanner.go`)
2. Implement the module struct and `Run()` method
3. Add error handling and timeout support
4. Register the module in `main.go`
5. Test the module independently

### Code Structure

Each module follows this pattern:
```go
type ModuleName struct {
    config *config.Config
}

func NewModuleName(cfg *config.Config) *ModuleName {
    return &ModuleName{config: cfg}
}

func (m *ModuleName) Run() error {
    // Implementation with error handling and timeout
}
```

## Testing

To test error handling (without feroxbuster installed):
```bash
./subdomain-recon https://subdomain.test.com
# Should skip module gracefully
```

With feroxbuster installed:
```bash
# Create a small test wordlist
echo -e "admin\napi\ntest" > /tmp/test-wordlist.txt

# Run with test wordlist (modify config first)
./subdomain-recon https://example.com
```

## License

This tool is for authorized security testing only. Always get permission before scanning any target.

## Contributing

This is a modular tool being built step-by-step. Each module is tested independently before moving to the next.
