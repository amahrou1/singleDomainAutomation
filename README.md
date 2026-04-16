# Single Subdomain Reconnaissance Tool

A modular Go-based reconnaissance tool focused on scanning a single subdomain for bug hunting and security testing.

## Version 2.0 - What's New! 🚀

- ✅ **YAML Configuration File** - Easy-to-edit `config.yaml` for all settings
- ✅ **Dirsearch Module** - Advanced path fuzzing with 74+ file extensions
- ✅ **Dual Content Discovery** - Run both Feroxbuster AND Dirsearch
- ✅ **Multiple Input Modes** - Single domain (`-d`) or batch scan from file (`-l`)
- ✅ **Custom Output Directory** - Use `-o /path/to/dir` to send all results into one location
- ✅ **Enhanced CLI** - Help flags, version info, custom config paths
- ✅ **Sequential Batch Processing** - Scan multiple targets one by one
- ✅ **Better Logging** - Module execution summary and success/failure tracking

## Features

### Current Modules

#### Module 1: Feroxbuster (Fast Directory/File Bruteforce)
- High-performance directory and file discovery
- Customizable status code filtering
- Configurable threading
- 5-hour timeout protection
- Output: `feroxbuster.txt`

#### Module 2: Dirsearch (Advanced Path Fuzzing)
- Comprehensive file extension coverage (74+ extensions)
- Intelligent path fuzzing
- Full URL output support
- Recursive scanning option
- Output: `dirsearch.txt`

## Project Structure

```
singleDomainAutomation/
├── main.go                          # Main entry point with CLI
├── config.yaml                      # User-editable configuration
├── go.mod                           # Go dependencies
├── config/
│   └── config.go                    # Configuration management
├── modules/
│   ├── content_discovery.go         # Feroxbuster module
│   └── dirsearch.go                 # Dirsearch module
├── utils/
│   └── logger.go                    # Logging utilities
├── output/                          # Output directory for results
│   ├── feroxbuster.txt             # Feroxbuster results
│   └── dirsearch.txt               # Dirsearch results
└── subdomain-recon                  # Compiled binary
```

## Prerequisites

### 1. Go (version 1.19 or higher)
```bash
go version
```

### 2. Feroxbuster
Fast content discovery tool written in Rust.

```bash
# Install on Linux
curl -sL https://raw.githubusercontent.com/epi052/feroxbuster/main/install-nix.sh | bash

# Or download from releases
wget https://github.com/epi052/feroxbuster/releases/latest/download/x86_64-linux-feroxbuster.tar.gz
tar -xzf x86_64-linux-feroxbuster.tar.gz
sudo mv feroxbuster /usr/local/bin/
chmod +x /usr/local/bin/feroxbuster

# Verify installation
feroxbuster --version
```

### 3. Dirsearch
Advanced web path scanner written in Python.

```bash
# Install with pip
pip3 install dirsearch

# Or clone from GitHub
git clone https://github.com/maurosoria/dirsearch.git
cd dirsearch
pip3 install -r requirements.txt
sudo ln -s $(pwd)/dirsearch.py /usr/local/bin/dirsearch

# Verify installation
dirsearch --version
```

### 4. Wordlist
Place your wordlist at `/root/myLists/all.txt` or configure a different path in `config.yaml`.

**Recommended wordlists:**
- [SecLists](https://github.com/danielmiessler/SecLists)
- [Assetnote Wordlists](https://wordlists.assetnote.io/)
- [FuzzDB](https://github.com/fuzzdb-project/fuzzdb)

## Installation

### Basic Installation

1. Clone or download this repository
   ```bash
   git clone <repository-url>
   cd singleDomainAutomation
   ```

2. Build the application
   ```bash
   go build -o subdomain-recon main.go
   ```

3. Make it executable
   ```bash
   chmod +x subdomain-recon
   ```

### System-Wide Installation (Ubuntu/Debian)

To run the tool from anywhere on your system:

**Option 1: Add to /usr/local/bin (Recommended)**
```bash
# Build the application
go build -o subdomain-recon main.go

# Move to /usr/local/bin
sudo mv subdomain-recon /usr/local/bin/

# Copy config file to your home directory
cp config.yaml ~/.subdomain-recon.yaml

# Verify installation
subdomain-recon -v
```

**Option 2: Add to PATH via ~/.bashrc**
```bash
# Build the application
go build -o subdomain-recon main.go

# Create a bin directory in your home folder
mkdir -p ~/bin

# Move the binary
mv subdomain-recon ~/bin/

# Copy config to home directory
cp config.yaml ~/.subdomain-recon.yaml

# Add to PATH in .bashrc
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc

# Reload .bashrc
source ~/.bashrc

# Verify installation
subdomain-recon -v
```

**Option 3: Create symbolic link**
```bash
# Build the application
go build -o subdomain-recon main.go

# Create symlink in /usr/local/bin
sudo ln -s $(pwd)/subdomain-recon /usr/local/bin/subdomain-recon

# Keep config.yaml in the project directory
# Use -c flag to specify config path when running from other directories

# Verify installation
subdomain-recon -v
```

**Using with custom config from anywhere:**
```bash
# If you installed system-wide, you can use a config file from anywhere
subdomain-recon -c /path/to/config.yaml https://target.com

# Or use the default config from home directory
mv config.yaml ~/.subdomain-recon.yaml
subdomain-recon -c ~/.subdomain-recon.yaml https://target.com
```

### Updating an Existing Installation (Ubuntu)

If you already have the tool installed and want to pull in the latest changes
(for example, the new `-o` output directory flag), run the following from the
cloned project directory:

```bash
# 1. Pull the latest code
cd /path/to/singleDomainAutomation
git pull origin main         # or: git pull origin <your-branch>

# 2. Rebuild the binary
go build -o subdomain-recon main.go

# 3a. If you installed to /usr/local/bin (Option 1):
sudo mv -f subdomain-recon /usr/local/bin/

# 3b. If you installed to ~/bin (Option 2):
mv -f subdomain-recon ~/bin/

# 3c. If you used a symlink (Option 3), no move needed — the symlink
#     already points at the freshly built binary in the repo.

# 4. Verify the new flag is available
subdomain-recon -h | grep -- '-o'
```

You should now see the `-o <dir>` option in the help output. Test it with:

```bash
subdomain-recon -l subdomains.txt -o /root/target/output
```

## Configuration

All settings are managed through `config.yaml`. Edit this file to customize:

### General Settings
```yaml
general:
  output_dir: "./output"    # NOTE: Ignored. Default output dir is ./fuzzing-output or override with -o flag
  timeout: "5h"             # Max timeout per module (5h, 30m, 2h30m)
```

**Note:** The output directory is `./fuzzing-output` by default. You can override
it with the `-o` flag on the command line, for example:
```bash
subdomain-recon -l subdomains.txt -o /root/target/output
```
All modules (feroxbuster, dirsearch, crawling) will write their single consolidated
result files inside that directory.

### Enable/Disable Modules
```yaml
modules:
  feroxbuster: true         # Enable feroxbuster
  dirsearch: true           # Enable dirsearch
```

### Feroxbuster Configuration
```yaml
feroxbuster:
  wordlist: "/root/myLists/all.txt"
  status_codes:             # HTTP codes to include
    - 200
    - 403
    - 301
    - 302
    - 307
  output_file: "feroxbuster.txt"
  threads: 50               # Concurrent threads
  recursion_depth: 0        # 0 = no recursion
  skip_ssl: true            # Skip SSL verification
  no_state: true            # Don't save/load state
```

### Dirsearch Configuration
```yaml
dirsearch:
  wordlist: ""              # Leave empty for default
  extensions:               # 74 file extensions configured
    - php
    - asp
    - aspx
    - jsp
    # ... (see config.yaml for full list)
  status_codes:
    - 200
    - 403
    - 301
    - 302
  output_file: "dirsearch.txt"
  full_url: true            # Show full URLs
  recursive: true           # Follow redirects
  threads: 30               # Concurrent threads
```

## Usage

### Input Modes

The tool supports three input modes for maximum flexibility:

#### 1. Single Domain Mode (`-d` flag)
Scan a single subdomain or IP address:
```bash
subdomain-recon -d https://api.example.com
subdomain-recon -d https://104.16.58.31
subdomain-recon -d https://104.16.58.31:8443
```

#### 2. Multiple Domains Mode (`-l` flag)
Scan multiple subdomains/IPs from a file:

**Create a file** (`targets.txt`):
```
# Subdomains
https://api.example.com
https://admin.example.com
https://dev.example.com

# IP addresses
https://104.16.58.31
https://104.16.58.31:8443
http://192.168.1.1:8080

# Comments are supported
https://staging.example.com
```

**Run the scan:**
```bash
subdomain-recon -l subdomains.txt
```

The tool will process each subdomain **sequentially**:
1. Scan `https://api.example.com` → Create `./api.example.com/` → Run all modules
2. Scan `https://admin.example.com` → Create `./admin.example.com/` → Run all modules
3. Scan `https://dev.example.com` → Create `./dev.example.com/` → Run all modules
4. And so on...

#### 3. Backward Compatible Mode (positional argument)
```bash
subdomain-recon https://subdomain.test.com
```

### Usage Examples

**Single domain scan:**
```bash
subdomain-recon -d https://api.test.com
```

**Multiple domains scan:**
```bash
subdomain-recon -l targets.txt
```

**With custom output directory (`-o` flag):**
```bash
# All dirsearch, feroxbuster, and crawling results will be written
# into /root/target/output (one consolidated file per module).
subdomain-recon -l targets.txt -o /root/target/output
subdomain-recon -d https://api.test.com -o /root/target/output
```

**With per-target timeout (`-t` flag):**
```bash
# Cap each module at 10 minutes per target. After 10 minutes the current
# module is killed (partial results are still saved) and the tool moves on
# to the next module / next target. Applies to feroxbuster, dirsearch, and
# crawling alike. Overrides the `timeout:` value in config.yaml.
subdomain-recon -l targets.txt -o /root/target/output -t 10m
```

**With custom config:**
```bash
subdomain-recon -d https://api.test.com -c ~/.subdomain-recon.yaml
subdomain-recon -l targets.txt -c ~/.subdomain-recon.yaml
```

**Show help:**
```bash
subdomain-recon -h
```

**Show version:**
```bash
subdomain-recon -v
```

**Run without building:**
```bash
go run main.go -d https://subdomain.test.com
go run main.go -l subdomains.txt
```

### IP Address Support

The tool fully supports scanning IP addresses with or without ports:

#### Supported Formats:
```bash
# IP without port (uses default: 443 for https, 80 for http)
subdomain-recon -d https://104.16.58.31
subdomain-recon -d http://192.168.1.1

# IP with custom port
subdomain-recon -d https://104.16.58.31:8443
subdomain-recon -d http://192.168.1.1:8080
```

#### Directory Naming for IPs:

**Standard ports (80 for HTTP, 443 for HTTPS):**
- `https://104.16.58.31` → `./104.16.58.31/`
- `https://104.16.58.31:443` → `./104.16.58.31/` (port omitted)
- `http://192.168.1.1:80` → `./192.168.1.1/` (port omitted)

**Non-standard ports:**
- `https://104.16.58.31:8443` → `./104.16.58.31_8443/` (port included with underscore)
- `http://192.168.1.1:8080` → `./192.168.1.1_8080/` (port included with underscore)

This ensures that scanning the same IP on different ports creates separate directories.

#### Mixed File Example:
Your targets file can contain both subdomains and IPs:
```
# Subdomains
https://api.example.com
https://admin.example.com

# IP addresses
https://104.16.58.31
https://104.16.58.31:8443
https://10.0.0.5:9443

# Mixed
http://192.168.1.100:8080
https://dev.example.com
```

## Command Details

### Feroxbuster Command
The tool executes:
```bash
feroxbuster -u https://subdomain.test.com \
  -w /root/myLists/all.txt \
  -s 200,403,301,302,307 \
  -t 50 \
  -k \
  -n \
  --no-state \
  -o <output-dir>/feroxbuster.txt
```

Where `<output-dir>` is `./fuzzing-output` by default, or whatever you pass to `-o`.

**Parameters:**
- `-u` : Target URL
- `-w` : Wordlist path
- `-s` : Status codes to include
- `-t` : Number of threads
- `-k` : Skip SSL certificate verification
- `-n` : Don't scan recursively
- `--no-state` : Don't save or load state
- `-o` : Output file path

### Dirsearch Command
The tool executes:
```bash
dirsearch -u https://subdomain.test.com \
  -e php,asp,aspx,jsp,py,txt,conf,... (74 extensions) \
  -i 200,403,301,302 \
  -t 30 \
  --full-url \
  -r \
  -o <output-dir>/dirsearch.txt
```

Where `<output-dir>` is `./fuzzing-output` by default, or whatever you pass to `-o`.

**Parameters:**
- `-u` : Target URL
- `-e` : File extensions to check
- `-i` : Status codes to include
- `-t` : Number of threads
- `--full-url` : Display full URLs in output
- `-r` : Recursive scanning (follow redirects)
- `-o` : Output file path

## Output

### Output Directory

By default, all results are saved to `./fuzzing-output/` in the current working
directory. You can override this location with the `-o` flag:

```bash
subdomain-recon -l subdomains.txt -o /root/target/output
```

### Output Files

Regardless of whether you scan a single domain or a batch from a file, results
are aggregated into **one consolidated file per module** inside the output
directory. Each target's results are appended with a clear header separator so
you can see what belongs to which subdomain/IP.

The output directory contains:
- `feroxbuster.txt` - Feroxbuster results for **all** targets (single file)
- `dirsearch.txt` - Dirsearch results for **all** targets (single file)
- `crawling-result.txt` - Crawling results for **all** targets (single file, live URLs only)

Each block in those files looks like:
```
================================================================================
[DIRSEARCH] Target: https://api.example.com
================================================================================
... results for this target ...

================================================================================
[DIRSEARCH] Target: https://admin.example.com
================================================================================
... results for this target ...
```

### Examples

```bash
# Default location (./fuzzing-output/)
subdomain-recon -l subdomains.txt
# Produces:
#   ./fuzzing-output/feroxbuster.txt
#   ./fuzzing-output/dirsearch.txt
#   ./fuzzing-output/crawling-result.txt

# Custom location via -o flag
subdomain-recon -l subdomains.txt -o /root/target/output
# Produces:
#   /root/target/output/feroxbuster.txt
#   /root/target/output/dirsearch.txt
#   /root/target/output/crawling-result.txt
```

## Error Handling

The tool includes robust error handling:

1. **Missing Tools**: Modules skip gracefully if tools aren't installed
2. **Missing Wordlist**: Module skips with clear error message
3. **Timeout Protection**: 5-hour timeout per module (configurable)
4. **Partial Results**: Saves results even on timeout or errors
5. **Continue on Failure**: Failed modules don't stop other modules

## Example Output

```
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║     Single Subdomain Reconnaissance Tool                 ║
║     Version: 2.0                                          ║
║     Modules: Feroxbuster + Dirsearch                      ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝

[INFO] Target: https://subdomain.test.com
[SUCCESS] Configuration loaded from: config.yaml

[Configuration]
─────────────────────────────────────────────────────────
Target:       https://subdomain.test.com
Output Dir:   ./output
Timeout:      5h0m0s (per module)

[Enabled Modules]
  Feroxbuster:  true
  Dirsearch:    true

[Feroxbuster Settings]
  Wordlist:     /root/myLists/all.txt
  Status Codes: [200 403 301 302 307]
  Threads:      50
  Output:       ./output/feroxbuster.txt

[Dirsearch Settings]
  Wordlist:     (default)
  Extensions:   74 configured
  Status Codes: [200 403 301 302]
  Threads:      30
  Output:       ./output/dirsearch.txt
─────────────────────────────────────────────────────────

============================================================
[INFO] Starting Module: Feroxbuster (Directory/File Bruteforce)
============================================================
[INFO] Running: feroxbuster -u https://subdomain.test.com ...
[INFO] Output will be saved to: ./output/feroxbuster.txt
[INFO] Timeout: 5h0m0s

[feroxbuster scan output...]

[SUCCESS] Module 'Feroxbuster' completed successfully
[SUCCESS] Output saved to: ./output/feroxbuster.txt (Size: 15234 bytes)
============================================================

============================================================
[INFO] Starting Module: Dirsearch (Advanced Path Fuzzing)
============================================================
[INFO] Running: dirsearch -u https://subdomain.test.com ...
[INFO] Output will be saved to: ./output/dirsearch.txt
[INFO] Timeout: 5h0m0s

[dirsearch scan output...]

[SUCCESS] Module 'Dirsearch' completed successfully
[SUCCESS] Output saved to: ./output/dirsearch.txt (Size: 28456 bytes)
============================================================

[Module Execution Summary]
─────────────────────────────────────────────────────────
Total Modules:      2
Successful:         2
Failed/Skipped:     0
─────────────────────────────────────────────────────────

╔═══════════════════════════════════════════════════════════╗
║                     Scan Complete                         ║
╚═══════════════════════════════════════════════════════════╝
Total Execution Time: 25m43s

Results Directory: ./output

💡 Tip: Review the output files for discovered paths and files
```

## Troubleshooting

### Tools Not Found
```
[WARNING] Module skipped: feroxbuster is not installed or not in PATH
```
**Solution:** Install feroxbuster and ensure it's in your PATH

### Wordlist Not Found
```
[WARNING] Module skipped: wordlist not found at /root/myLists/all.txt
```
**Solution:** Update the wordlist path in `config.yaml` or create the wordlist at the specified location

### Configuration Error
```
[ERROR] Failed to load configuration: invalid timeout format '5hours'
```
**Solution:** Use correct time format in `config.yaml` (e.g., "5h", "30m", "2h30m")

### Timeout Issues
If scans timeout frequently:
1. Increase timeout in `config.yaml`: `timeout: "10h"`
2. Reduce thread count to avoid rate limiting
3. Use smaller wordlists for initial scans

## Future Modules (Planned)

- Module 3: Port Scanning (nmap/masscan)
- Module 4: Technology Detection (whatweb/wappalyzer)
- Module 5: SSL/TLS Analysis (testssl.sh)
- Module 6: JavaScript Analysis (subjs/linkfinder)
- Module 7: Parameter Discovery (arjun)
- Module 8: Subdomain Takeover Check

## Development

### Building from Source
```bash
go mod tidy
go build -o subdomain-recon main.go
```

### Adding New Modules

1. Create a new file in `modules/` (e.g., `port_scanner.go`)
2. Implement the module struct and `Run()` method following the existing pattern
3. Add module configuration to `config.yaml`
4. Update `config/config.go` with module-specific config struct
5. Register the module in `main.go` `runModules()` function
6. Test independently before integration
7. Update this README

### Module Template
```go
type NewModule struct {
    config *config.Config
}

func NewNewModule(cfg *config.Config) *NewModule {
    return &NewModule{config: cfg}
}

func (m *NewModule) Run() error {
    moduleName := "New Module Name"
    utils.LogModuleStart(moduleName)
    startTime := time.Now()

    // Check prerequisites
    // Run module logic with timeout
    // Handle errors gracefully

    duration := time.Since(startTime)
    utils.LogModuleEnd(moduleName, err, duration)
    return err
}
```

## Performance Tips

1. **Adjust Thread Count**: Higher threads = faster scans but more resource usage
2. **Use Focused Wordlists**: Smaller, targeted wordlists for specific technologies
3. **Disable Unused Modules**: Set `feroxbuster: false` or `dirsearch: false` in config
4. **Run Modules Separately**: For very large targets, run one module at a time
5. **Monitor Resource Usage**: Watch CPU/memory during scans

## Security & Legal

⚠️ **IMPORTANT**: This tool is for authorized security testing only.

- Always get explicit permission before scanning any target
- Respect rate limits and robots.txt
- Be aware of your country's computer misuse laws
- Use responsibly in bug bounty programs
- Don't use for malicious purposes

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This tool is for authorized security testing only.

## Credits

- **Feroxbuster**: [epi052/feroxbuster](https://github.com/epi052/feroxbuster)
- **Dirsearch**: [maurosoria/dirsearch](https://github.com/maurosoria/dirsearch)

## Support

For issues, questions, or feature requests:
- Open an issue on GitHub
- Check existing issues for solutions
- Provide detailed information (OS, Go version, error messages)

---

**Version**: 2.0
**Last Updated**: 2025-12-15
**Status**: Active Development
