# VTEX Files Manager (vfm)

[![npm version](https://img.shields.io/npm/v/vtex-files-manager.svg)](https://www.npmjs.com/package/vtex-files-manager)
[![License](https://img.shields.io/github/license/glinharesb/vtex-files-manager)](https://github.com/glinharesb/vtex-files-manager/blob/main/LICENSE)

A powerful CLI tool for uploading images and files to VTEX, integrated with VTEX CLI.

## Installation

```bash
npm install -g vtex-files-manager
```

The `vfm` command will be available globally after installation.

## Prerequisites

- [VTEX CLI](https://developers.vtex.com/docs/guides/vtex-io-documentation-vtex-io-cli-install) installed
- Authenticated session in VTEX CLI (`vtex login`)

## Quick Start

```bash
# Authenticate with VTEX CLI first
vtex login

# Upload a single file
vfm upload image.jpg -m cms

# Upload multiple files
vfm batch ./images -m graphql

# View upload history
vfm logs

# Update to latest version
vfm update

# Get help
vfm --help
```

## Features

- ✅ Single file or batch upload
- ✅ Automatic integration with VTEX CLI session
- ✅ Two upload methods: GraphQL (official) and CMS FilePicker (legacy)
- ✅ Automatic file existence verification
- ✅ Confirmation prompts before overwriting
- ✅ URL encoding for special characters
- ✅ Configurable concurrent uploads
- ✅ Recursive subdirectory support
- ✅ Progress bars during upload
- ✅ Upload history with logs command
- ✅ Self-update capability

## Updating

```bash
# Update via npm
npm update -g vtex-files-manager

# Or use the built-in update command
vfm update
```

## Documentation

For full documentation, visit: [https://github.com/glinharesb/vtex-files-manager](https://github.com/glinharesb/vtex-files-manager)

## Supported Formats

**Universal (both methods):** JPG, JPEG, PNG, GIF, SVG, WEBP
**CMS only:** BMP, PDF, TXT, JSON, XML, CSS, JS

**Maximum file size:** 5MB per file

## Troubleshooting

### Binary not found after installation

If you get "vfm: command not found" after installation:

```bash
# Reinstall the package
npm uninstall -g vtex-files-manager
npm install -g vtex-files-manager
```

### Authentication errors

Make sure you're logged in with VTEX CLI:

```bash
vtex login
```

## License

MIT License - see [LICENSE](https://github.com/glinharesb/vtex-files-manager/blob/main/LICENSE) for details.

## Links

- [GitHub Repository](https://github.com/glinharesb/vtex-files-manager)
- [Report Issues](https://github.com/glinharesb/vtex-files-manager/issues)
- [VTEX Documentation](https://developers.vtex.com/)
- [VTEX CLI](https://github.com/vtex/toolbelt)
