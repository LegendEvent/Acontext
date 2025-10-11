import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

// 获取当前文件的目录路径 (ES modules 没有 __dirname)
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// 上级目录路径
const parentDir = path.resolve(__dirname, '../../');
// 当前目录路径
const currentDir = path.resolve(__dirname, '../');

console.log('🔄 Syncing environment files from parent directory...');

try {
  // 读取上级目录的所有文件
  const files = fs.readdirSync(parentDir);

  // 筛选出所有 .env 开头的文件
  const envFiles = files.filter(file => file.startsWith('.env'));

  if (envFiles.length === 0) {
    console.log('⚠️  No .env files found in parent directory');
    process.exit(0);
  }

  // 复制所有 .env 文件到当前目录
  envFiles.forEach(file => {
    const sourcePath = path.join(parentDir, file);
    const targetPath = path.join(currentDir, file);

    fs.copyFileSync(sourcePath, targetPath);
    console.log(`✅ Copied ${file}`);
  });

  console.log(`✨ Successfully synced ${envFiles.length} environment file(s)`);
} catch (error) {
  console.error('❌ Error syncing environment files:', error.message);
  process.exit(1);
}

