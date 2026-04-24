#!/usr/bin/env node

const path = require('path');
const { spawn } = require('child_process');
const os = require('os');

const platform = os.platform();
const binDir = path.join(__dirname, '..', 'bin');

let binary;
if (platform === 'win32') {
  binary = path.join(binDir, 'cbi.exe');
} else {
  binary = path.join(binDir, 'cbi');
}

// 运行二进制文件，传递所有参数
const args = process.argv.slice(2);
const child = spawn(binary, args, { stdio: 'inherit' });

child.on('exit', (code) => {
  process.exit(code || 0);
});