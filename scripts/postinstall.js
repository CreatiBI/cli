#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const os = require('os');
const { execSync } = require('child_process');

// 获取当前平台信息
const platform = os.platform();
const arch = os.arch();

// 映射到我们的压缩包路径
const archiveMap = {
  'darwin-x64': { archive: 'bin/darwin-amd64/cbi.tar.gz', binary: 'cbi' },
  'darwin-arm64': { archive: 'bin/darwin-arm64/cbi.tar.gz', binary: 'cbi' },
  'linux-x64': { archive: 'bin/linux-amd64/cbi.tar.gz', binary: 'cbi' },
  'linux-arm64': { archive: 'bin/linux-arm64/cbi.tar.gz', binary: 'cbi' },
  'win32-x64': { archive: 'bin/windows-amd64/cbi.zip', binary: 'cbi.exe' },
  'win32-arm64': { archive: 'bin/windows-arm64/cbi.zip', binary: 'cbi.exe' }
};

const key = `${platform}-${arch}`;
const config = archiveMap[key];

if (!config) {
  console.error(`Unsupported platform: ${platform}-${arch}`);
  process.exit(1);
}

const binDir = path.join(__dirname, '..', 'bin');
const archivePath = path.join(__dirname, '..', config.archive);
const targetPath = path.join(binDir, 'cbi');
const targetBinary = platform === 'win32' ? 'cbi.exe' : 'cbi';

// 检查压缩包是否存在
if (!fs.existsSync(archivePath)) {
  console.error(`Archive not found: ${archivePath}`);
  console.error('Please run: npm run prepack');
  process.exit(1);
}

// 确保目标目录存在
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

// 解压文件
console.log(`Extracting binary for ${platform}-${arch}...`);

try {
  if (platform === 'win32') {
    // Windows 使用 PowerShell 解压
    execSync(`powershell -command "Expand-Archive -Path '${archivePath}' -DestinationPath '${binDir}' -Force"`, { stdio: 'inherit' });
  } else {
    // macOS/Linux 使用 tar 解压
    execSync(`tar -xzf "${archivePath}" -C "${binDir}"`, { stdio: 'inherit' });
  }
} catch (err) {
  console.error('Failed to extract archive:', err.message);
  process.exit(1);
}

// 检查解压后的文件
const extractedPath = path.join(binDir, targetBinary);
if (!fs.existsSync(extractedPath)) {
  console.error(`Binary not found after extraction: ${extractedPath}`);
  process.exit(1);
}

// 重命名为 cbi（Windows 需要）
if (platform === 'win32') {
  // Windows 已解压为 cbi.exe，需要确保 bin/cbi.cmd 存在
  const cmdPath = path.join(binDir, 'cbi.cmd');
  fs.writeFileSync(cmdPath, `@echo off\n"%~dp0cbi.exe" %*\n`);
} else {
  // 设置可执行权限
  fs.chmodSync(extractedPath, 0o755);
}

console.log(`✓ Installed cbi for ${platform}-${arch}`);