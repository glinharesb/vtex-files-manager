#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

// Determine binary name based on platform
const platform = process.platform;
const binaryName = platform === 'win32' ? 'vfm.exe' : 'vfm';
const binaryPath = path.join(__dirname, binaryName);

// Check if binary exists
if (!fs.existsSync(binaryPath)) {
  console.error('Error: vfm binary not found.');
  console.error('Please reinstall the package: npm install -g vtex-files-manager');
  process.exit(1);
}

// Execute the binary with all arguments
const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit'
});

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
  } else {
    process.exit(code);
  }
});

child.on('error', (err) => {
  console.error('Failed to execute vfm:', err.message);
  process.exit(1);
});
