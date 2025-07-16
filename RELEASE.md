# 发布指南

本文档说明如何使用新添加的自动化构建和发布系统。

## 🚀 首次发布

### 1. 推送 Tag 触发自动发布

```bash
# 创建并推送 tag
git tag v1.0.0
git push origin v1.0.0
```

### 2. 手动触发发布 (可选)

也可以在 GitHub 仓库的 Actions 页面手动触发 Release 工作流：

1. 访问 GitHub 仓库
2. 点击 "Actions" 标签
3. 选择 "Release" 工作流
4. 点击 "Run workflow"
5. 输入版本号 (如 `v1.0.0`)
6. 点击 "Run workflow"

## 📦 发布内容

每次发布会自动生成以下文件：

- `alex-linux-amd64` - Linux x64 二进制文件
- `alex-linux-arm64` - Linux ARM64 二进制文件  
- `alex-darwin-amd64` - macOS Intel 二进制文件
- `alex-darwin-arm64` - macOS Apple Silicon 二进制文件
- `alex-windows-amd64.exe` - Windows x64 二进制文件
- `checksums.txt` - SHA256 校验文件

## 🛠️ 安装方式

发布后，用户可以通过以下方式安装：

### 快速安装 (推荐)

**Linux/macOS:**
```bash
curl -sSfL https://raw.githubusercontent.com/ckl/Alex-Code/main/scripts/install.sh | sh
```

**Windows:**
```powershell
iwr -useb https://raw.githubusercontent.com/ckl/Alex-Code/main/scripts/install.ps1 | iex
```

### 手动下载

用户也可以直接从 Releases 页面下载对应平台的二进制文件。

## ⚙️ 配置说明

### 修改 GitHub 仓库路径

在发布前，请确保修改以下文件中的仓库路径：

1. **安装脚本中的仓库路径:**
   - `scripts/install.sh` 第9行: `GITHUB_REPO="ckl/Alex-Code"`
   - `scripts/install.ps1` 第6行: `[string]$Repository = "ckl/Alex-Code"`

2. **文档中的链接:**
   - `docs/installation.md` 中的所有GitHub链接

### 版本号格式

建议使用语义化版本号格式：
- `v1.0.0` - 主要版本
- `v1.1.0` - 次要版本  
- `v1.1.1` - 补丁版本

## 🔍 验证发布

发布完成后，可以通过以下方式验证：

1. **检查 Releases 页面:**
   - 确认所有平台的二进制文件都已生成
   - 确认 checksums.txt 文件存在

2. **测试安装脚本:**
   ```bash
   # 测试Linux/macOS安装脚本
   ./scripts/install.sh --version v1.0.0 --repo your-org/your-repo
   
   # 测试Windows安装脚本
   .\scripts\install.ps1 -Version v1.0.0 -Repository "your-org/your-repo"
   ```

3. **验证二进制文件:**
   ```bash
   # 下载并测试二进制文件
   alex --version
   alex --help
   ```

## 📋 发布检查清单

在发布前请确认：

- [ ] 代码已提交并推送到主分支
- [ ] 版本号已在代码中更新 (如果需要)
- [ ] 安装脚本中的仓库路径已正确配置
- [ ] 文档中的链接已更新为正确的仓库路径
- [ ] 已测试主要功能正常工作
- [ ] 准备好发布说明 (GitHub会自动生成)

## 🐛 故障排除

### 构建失败

如果 GitHub Actions 构建失败：

1. 检查 Actions 页面的错误日志
2. 确认 Go 版本兼容性
3. 检查依赖项是否正确
4. 验证 LDFLAGS 是否正确设置

### 发布失败  

如果发布过程失败：

1. 确认 GITHUB_TOKEN 权限正确
2. 检查 tag 格式是否正确
3. 确认没有重复的 tag

### 安装脚本问题

如果用户反馈安装问题：

1. 检查二进制文件是否正确生成
2. 验证下载链接是否有效
3. 确认文件权限设置正确

## 📞 支持

如果遇到问题，可以：

1. 查看 GitHub Actions 日志
2. 检查 Issues 页面的类似问题
3. 创建新的 Issue 寻求帮助 