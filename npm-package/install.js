#!/usr/bin/env node

const https = require('https');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const zlib = require('zlib');
const tar = require('tar');

const REPO_OWNER = 'glinharesb';
const REPO_NAME = 'vtex-files-manager';
const BINARY_NAME = 'vfm';

// Detect platform and architecture
function getPlatform() {
  const platform = process.platform;
  const arch = process.arch;

  let os, architecture, ext, archivExt;

  // Map Node.js platform to GoReleaser naming
  switch (platform) {
    case 'darwin':
      os = 'Darwin';
      ext = '';
      archiveExt = '.tar.gz';
      break;
    case 'linux':
      os = 'Linux';
      ext = '';
      archiveExt = '.tar.gz';
      break;
    case 'win32':
      os = 'Windows';
      ext = '.exe';
      archiveExt = '.zip';
      break;
    default:
      throw new Error(`Unsupported platform: ${platform}`);
  }

  // Map Node.js arch to GoReleaser naming
  switch (arch) {
    case 'x64':
      architecture = 'x86_64';
      break;
    case 'arm64':
      architecture = 'arm64';
      break;
    default:
      throw new Error(`Unsupported architecture: ${arch}`);
  }

  return { os, architecture, ext, archiveExt };
}

// Fetch latest release information from GitHub API
async function getLatestRelease() {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'api.github.com',
      path: `/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest`,
      method: 'GET',
      headers: {
        'User-Agent': 'vtex-files-manager-npm-installer'
      }
    };

    https.get(options, (res) => {
      let data = '';

      res.on('data', (chunk) => {
        data += chunk;
      });

      res.on('end', () => {
        if (res.statusCode === 200) {
          resolve(JSON.parse(data));
        } else {
          reject(new Error(`Failed to fetch latest release: HTTP ${res.statusCode}`));
        }
      });
    }).on('error', reject);
  });
}

// Download file from URL
async function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);

    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Follow redirect
        return downloadFile(response.headers.location, dest).then(resolve).catch(reject);
      }

      if (response.statusCode !== 200) {
        reject(new Error(`Failed to download: HTTP ${response.statusCode}`));
        return;
      }

      response.pipe(file);

      file.on('finish', () => {
        file.close();
        resolve();
      });
    }).on('error', (err) => {
      fs.unlink(dest, () => {});
      reject(err);
    });
  });
}

// Extract tar.gz archive
async function extractTarGz(archivePath, destDir) {
  return tar.x({
    file: archivePath,
    cwd: destDir
  });
}

// Extract zip archive (Windows)
async function extractZip(archivePath, destDir) {
  // Use unzip command or built-in extraction
  try {
    execSync(`tar -xf "${archivePath}" -C "${destDir}"`, { stdio: 'ignore' });
  } catch (err) {
    // Fallback for Windows - try PowerShell
    const psCommand = `Expand-Archive -Path "${archivePath}" -DestinationPath "${destDir}" -Force`;
    execSync(`powershell -Command "${psCommand}"`, { stdio: 'ignore' });
  }
}

// Main installation function
async function install() {
  try {
    console.log('📦 Installing vtex-files-manager (vfm)...');

    // Detect platform
    const { os, architecture, ext, archiveExt } = getPlatform();
    console.log(`🔍 Detected platform: ${os} ${architecture}`);

    // Get latest release
    console.log('🔍 Fetching latest release...');
    const release = await getLatestRelease();
    const version = release.tag_name.replace(/^v/, '');
    console.log(`✓ Latest version: ${release.tag_name}`);

    // Construct download URL based on GoReleaser naming convention
    const archiveName = `${REPO_NAME}_${version}_${os}_${architecture}${archiveExt}`;
    const downloadUrl = `https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${release.tag_name}/${archiveName}`;

    console.log(`⬇️  Downloading binary from GitHub...`);

    // Create temp directory
    const tempDir = path.join(__dirname, '.tmp');
    if (!fs.existsSync(tempDir)) {
      fs.mkdirSync(tempDir, { recursive: true });
    }

    // Download archive
    const archivePath = path.join(tempDir, archiveName);
    await downloadFile(downloadUrl, archivePath);
    console.log('✓ Download complete');

    // Extract archive
    console.log('📂 Extracting binary...');
    if (archiveExt === '.tar.gz') {
      await extractTarGz(archivePath, tempDir);
    } else {
      await extractZip(archivePath, tempDir);
    }

    // Move binary to bin directory
    const binDir = path.join(__dirname, 'bin');
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }

    const sourceBinary = path.join(tempDir, BINARY_NAME + ext);
    const destBinary = path.join(binDir, BINARY_NAME + ext);

    if (fs.existsSync(destBinary)) {
      fs.unlinkSync(destBinary);
    }

    fs.renameSync(sourceBinary, destBinary);

    // Make executable on Unix systems
    if (os !== 'Windows') {
      fs.chmodSync(destBinary, 0o755);
    }

    // Clean up
    fs.rmSync(tempDir, { recursive: true, force: true });

    console.log('✅ Installation complete!');
    console.log(`\n🚀 You can now use: vfm --help\n`);

  } catch (error) {
    console.error('❌ Installation failed:', error.message);
    console.error('\n📖 Please try manual installation:');
    console.error(`   https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest\n`);
    process.exit(1);
  }
}

// Run installation
install();
