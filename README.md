# VTEX Files Manager

[![CI](https://github.com/glinharesb/vtex-files-manager/workflows/CI/badge.svg)](https://github.com/glinharesb/vtex-files-manager/actions?query=workflow%3ACI)
[![Release](https://github.com/glinharesb/vtex-files-manager/workflows/Release/badge.svg)](https://github.com/glinharesb/vtex-files-manager/actions?query=workflow%3ARelease)
[![Go Version](https://img.shields.io/github/go-mod/go-version/glinharesb/vtex-files-manager)](https://github.com/glinharesb/vtex-files-manager/blob/main/go.mod)
[![License](https://img.shields.io/github/license/glinharesb/vtex-files-manager)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/glinharesb/vtex-files-manager)](https://github.com/glinharesb/vtex-files-manager/releases/latest)

**vfm** - A powerful CLI tool for uploading images and files to VTEX, integrated with VTEX CLI.

## Features

- ✅ Single file or batch upload
- ✅ Automatic integration with VTEX CLI (uses `vtex login` session)
- ✅ Two upload methods: GraphQL (official) and CMS FilePicker (legacy)
- ✅ Automatic check for existing files
- ✅ Confirmation prompt before overwriting
- ✅ URL encoding for names with spaces and special characters
- ✅ Configurable concurrent uploads
- ✅ Recursive subdirectory support
- ✅ Progress bar during upload
- ✅ Upload history with logs command

## Prerequisites

- [VTEX CLI](https://developers.vtex.com/docs/guides/vtex-io-documentation-vtex-io-cli-install) installed
- Authenticated session in VTEX CLI (`vtex login`)

## Installation

### Option 1: npm (Recommended)

```bash
npm install -g vtex-files-manager
```

The `vfm` command will be available globally.

### Option 2: Direct Binary Download

Download the `vfm` binary for your operating system from the [releases page](https://github.com/glinharesb/vtex-files-manager/releases/latest):

#### **Linux**
```bash
# Download (replace VERSION with desired version, e.g., 1.0.0)
wget https://github.com/glinharesb/vtex-files-manager/releases/download/vVERSION/vtex-files-manager_VERSION_Linux_x86_64.tar.gz

# Extract
tar -xzf vtex-files-manager_VERSION_Linux_x86_64.tar.gz

# Move to PATH (requires sudo)
sudo mv vfm /usr/local/bin/

# Test
vfm --help
```

#### **macOS**
```bash
# Intel (x86_64)
wget https://github.com/glinharesb/vtex-files-manager/releases/download/vVERSION/vtex-files-manager_VERSION_Darwin_x86_64.tar.gz
tar -xzf vtex-files-manager_VERSION_Darwin_x86_64.tar.gz

# Apple Silicon (M1/M2/M3)
wget https://github.com/glinharesb/vtex-files-manager/releases/download/vVERSION/vtex-files-manager_VERSION_Darwin_arm64.tar.gz
tar -xzf vtex-files-manager_VERSION_Darwin_arm64.tar.gz

# Move to PATH
sudo mv vfm /usr/local/bin/

# Test
vfm --help
```

#### **Windows**
1. Download the appropriate `.zip` file from the [releases page](https://github.com/glinharesb/vtex-files-manager/releases/latest)
2. Extract the `vfm.exe` file
3. Move to a directory in your PATH (e.g., `C:\Program Files\vfm\`)
4. Add the directory to system PATH
5. Open a new terminal and test: `vfm --help`

### Option 3: Install via Go

```bash
go install github.com/glinharesb/vtex-files-manager@latest
```

The binary will be installed as `vtex-files-manager` in `$GOPATH/bin`. You can create an alias `vfm` or rename the binary.

### Option 4: Build from Source

```bash
git clone https://github.com/glinharesb/vtex-files-manager.git
cd vtex-files-manager

# Build vfm
go build -o vfm .

# Optional: Move to PATH
sudo mv vfm /usr/local/bin/

# Test
vfm --help
```

## Usage

### Authentication

The tool automatically uses the VTEX CLI session. Make sure you're logged in:

```bash
vtex login
```

### Single File Upload

```bash
vfm upload <file> -m <method>
```

**Examples:**
```bash
# Upload with CMS FilePicker (short URLs)
vfm upload image.jpg -m cms

# Upload with GraphQL (official)
vfm upload logo.png -m graphql

# Skip confirmation prompt
vfm upload banner.jpg -m cms -y
```

### Batch Upload

```bash
vfm batch <directory> -m <method> [flags]
```

**Examples:**
```bash
# Upload all files in a directory
vfm batch ./images -m cms

# Recursive upload with 5 concurrent workers
vfm batch ./assets -m graphql -r -c 5

# Direct batch (no confirmation)
vfm batch ./photos -m cms -y
```

### View Upload Logs

```bash
vfm logs [flags]
```

**Examples:**
```bash
# View last 50 uploads (default)
vfm logs

# View only last 10
vfm logs --limit 10

# View only failed uploads
vfm logs --status failed

# View only CMS uploads
vfm logs --method cms

# Combine filters
vfm logs --status success --method graphql --limit 20

# Clear all logs
vfm logs --clear
```

The logs command displays:
- Upload timestamp
- File name and size
- Method used (CMS or GraphQL)
- Account and workspace
- Status (success or failure)
- Generated URL (if success)
- Error message (if failure)
- Summary statistics

**Log location:**
- Linux: `~/.local/state/vtex-files-manager/uploads.jsonl`
- macOS: `~/Library/Application Support/vtex-files-manager/uploads.jsonl`
- Windows: `%LOCALAPPDATA%\vtex-files-manager\uploads.jsonl`

## Upload Methods

### CMS FilePicker (`-m cms`)
- **Advantage**: Short and predictable URLs
- **URL**: `https://{account}.vtexassets.com/arquivos/filename.ext`
- **Use**: Upload via CMS admin (legacy)
- **Verification**: Detects existing files before overwriting

### GraphQL (`-m graphql`)
- **Advantage**: Official and modern API
- **URL**: `https://{account}.vtexassets.com/assets/.../uuid___hash.ext`
- **Use**: Upload via GraphQL mutation
- **Names**: Automatically generated (UUID + hash)

## Flags

### Upload Command

| Flag | Short | Description | Required |
|------|-------|-------------|----------|
| `--method` | `-m` | Upload method (cms or graphql) | ✅ |
| `--yes` | `-y` | Skip confirmation prompt | ❌ |
| `--verbose` | `-v` | Verbose output | ❌ |

### Batch Command

| Flag | Short | Description | Default | Required |
|------|-------|-------------|---------|----------|
| `--method` | `-m` | Upload method (cms or graphql) | - | ✅ |
| `--concurrent` | `-c` | Number of concurrent workers | 3 | ❌ |
| `--recursive` | `-r` | Search in subdirectories | false | ❌ |
| `--yes` | `-y` | Skip confirmation prompt | false | ❌ |
| `--verbose` | `-v` | Verbose output | false | ❌ |

### Logs Command

| Flag | Short | Description | Default | Required |
|------|-------|-------------|---------|----------|
| `--limit` | `-l` | Maximum entries to display | 50 | ❌ |
| `--status` | `-s` | Filter by status (success or failed) | - | ❌ |
| `--method` | `-m` | Filter by method (graphql or cms) | - | ❌ |
| `--clear` | `-c` | Clear all logs (requires confirmation) | false | ❌ |

## Supported Formats

The formats below have been validated against the real VTEX API:

| Format | CMS FilePicker | GraphQL | Category | Recommended Use |
|--------|----------------|---------|----------|-----------------|
| JPG/JPEG | ✅ | ✅ | Image | Universal |
| PNG | ✅ | ✅ | Image | Universal |
| GIF | ✅ | ✅ | Image | Universal |
| SVG | ✅ | ✅ | Image | Universal |
| WEBP | ✅ | ✅ | Image | Universal |
| BMP | ✅ | ❌ | Image | CMS only |
| PDF | ✅ | ❌ | Document | CMS only |
| TXT | ✅ | ❌ | Document | CMS only |
| JSON | ✅ | ❌ | Document | CMS only |
| XML | ✅ | ❌ | Document | CMS only |
| CSS | ✅ | ❌ | Web | CMS only |
| JS | ✅ | ❌ | Web | CMS only |

**Notes:**
- ✅ = Format accepted by API
- ❌ = Format rejected by API (returns "Invalid file format")
- **Universal**: Works with both methods (CMS and GraphQL)
- **CMS only**: Works only with CMS FilePicker method
- **Limit**: 5MB per file (all formats)

## Advanced Examples

### Upload with Existing File Confirmation

```bash
$ vfm upload image.jpg -m cms

=== VTEX File Upload ===
Account:       myaccount
Workspace:     master
User:          user@example.com
Method:        cms
File:          image.jpg (245 KB)
Destination:   https://myaccount.vtexassets.com/arquivos/image.jpg

⚠️  WARNING: File already exists and will be OVERWRITTEN!

File exists. Overwrite? [y/N]:
```

### Batch with Multiple File Verification

```bash
$ vfm batch ./images -m cms

=== VTEX Batch Upload ===
Account:       myaccount
Files found:   10 (5.2 MB total)

⚠️  WARNING: 3 file(s) already exist and will be OVERWRITTEN:
  • logo.png
  • banner.jpg
  • icon.svg

3 file(s) will be overwritten. Continue? [y/N]:
```

### Files with Spaces and Special Characters

The tool automatically handles URL encoding:

```bash
$ vfm upload "my file & photo.jpg" -m cms -y

✓ Upload successful!
File URL: https://myaccount.vtexassets.com/arquivos/my%20file%20&%20photo.jpg
```

## Project Structure

```
vtex-files-manager/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── upload.go          # Single upload command
│   ├── batch.go           # Batch upload command
│   ├── logs.go            # Log viewing command
│   └── helpers.go         # Shared helper functions
├── pkg/
│   ├── auth/              # Authentication
│   │   └── auth.go
│   ├── client/            # Upload clients
│   │   ├── common.go      # Shared code
│   │   ├── filepicker.go  # CMS FilePicker client
│   │   └── graphql.go     # GraphQL client
│   ├── logger/            # Logging system
│   │   └── upload_logger.go
│   └── vtexcli/           # VTEX CLI integration
│       └── session.go
└── main.go
```

## Troubleshooting

### Error: "No VTEX session found"

**Solution**: Run `vtex login` to authenticate.

### Error: "Failed to get requestToken"

**Solution**: Make sure your session hasn't expired. Run `vtex login` again.

### Slow Upload

**Solution**: For batch uploads, increase the number of workers with `-c`:
```bash
vfm batch ./images -m graphql -c 10
```

## Development

### Environment Setup

```bash
# Clone the repository
git clone https://github.com/glinharesb/vtex-files-manager.git
cd vtex-files-manager

# Download dependencies
go mod download

# Run tests
go test ./...

# Build vfm
go build -o vfm .

# Run
./vfm --help
```

### Code Structure

```
vtex-files-manager/
├── .github/
│   └── workflows/        # CI/CD workflows
├── cmd/                  # CLI commands (vfm)
│   ├── root.go          # Root command
│   ├── upload.go        # Single upload
│   ├── batch.go         # Batch upload
│   ├── logs.go          # Log viewing
│   └── helpers.go       # Helper functions
├── pkg/
│   ├── auth/            # Authentication
│   ├── client/          # Upload clients
│   │   ├── common.go    # Shared code
│   │   ├── filepicker.go # CMS FilePicker
│   │   └── graphql.go   # GraphQL API
│   ├── logger/          # Logging system
│   └── vtexcli/         # VTEX CLI integration
├── scripts/
│   └── release.sh       # Release script
├── .goreleaser.yml      # GoReleaser configuration
└── main.go
```

### Running Tests

```bash
# All tests
go test ./...

# With verbose output
go test -v ./...

# With coverage
go test -cover ./...

# Detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Creating a Release

```bash
# Use the automated script
./scripts/release.sh v1.2.3

# Or manually (see RELEASE.md)
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

See [RELEASE.md](RELEASE.md) for detailed instructions.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a branch for your feature (`git checkout -b feature/MyFeature`)
3. Commit your changes (`git commit -m 'Add: My feature'`)
4. Add tests if applicable
5. Make sure tests pass (`go test ./...`)
6. Push to the branch (`git push origin feature/MyFeature`)
7. Open a Pull Request

### Commit Conventions

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `test:` Tests
- `refactor:` Refactoring
- `style:` Formatting
- `chore:` Maintenance

**Examples:**
```bash
git commit -m "feat: add AVIF format support"
git commit -m "fix: fix timeout on large uploads"
git commit -m "docs: update README with new examples"
```

## Updating

### npm
```bash
npm update -g vtex-files-manager
```

### Direct download
```bash
vfm update
```

## License

MIT License - see LICENSE for details.

## Author

Gabriel Linhares Bernardes

## Links

- [VTEX Documentation](https://developers.vtex.com/)
- [VTEX CLI](https://github.com/vtex/toolbelt)
