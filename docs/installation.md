# Alex CLI 安装指南

Alex CLI 是一个强大的软件工程助手工具，支持多种平台和安装方式。

## 🚀 快速安装

### Linux/macOS

使用 curl 一键安装：

```bash
curl -sSfL https://raw.githubusercontent.com/ckl/Alex-Code/main/scripts/install.sh | sh
```

或者下载脚本后执行：

```bash
wget https://raw.githubusercontent.com/ckl/Alex-Code/main/scripts/install.sh
chmod +x install.sh
./install.sh
```

### Windows

使用 PowerShell 安装：

```powershell
# 如果需要，先设置执行策略
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# 下载并运行安装脚本
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ckl/Alex-Code/main/scripts/install.ps1" -OutFile "install.ps1"
.\install.ps1
```

或者一行命令：

```powershell
iwr -useb https://raw.githubusercontent.com/ckl/Alex-Code/main/scripts/install.ps1 | iex
```

## 📦 手动安装

### 1. 下载预编译二进制文件

访问 [Releases 页面](https://github.com/ckl/Alex-Code/releases/latest) 下载适合你系统的二进制文件：

| 平台 | 架构 | 文件名 |
|------|------|--------|
| Linux | x64 | `alex-linux-amd64` |
| Linux | ARM64 | `alex-linux-arm64` |
| macOS | Intel | `alex-darwin-amd64` |
| macOS | Apple Silicon | `alex-darwin-arm64` |
| Windows | x64 | `alex-windows-amd64.exe` |

### 2. 安装到系统

#### Linux/macOS

```bash
# 下载二进制文件 (以 Linux x64 为例)
wget https://github.com/ckl/Alex-Code/releases/latest/download/alex-linux-amd64

# 重命名并设置可执行权限
mv alex-linux-amd64 alex
chmod +x alex

# 移动到 PATH 目录
sudo mv alex /usr/local/bin/

# 或者移动到用户目录
mkdir -p ~/.local/bin
mv alex ~/.local/bin/
export PATH="$PATH:$HOME/.local/bin"
```

#### Windows

```powershell
# 下载二进制文件
Invoke-WebRequest -Uri "https://github.com/ckl/Alex-Code/releases/latest/download/alex-windows-amd64.exe" -OutFile "alex.exe"

# 创建安装目录
$installDir = "$env:LOCALAPPDATA\Alex"
New-Item -ItemType Directory -Path $installDir -Force

# 移动到安装目录
Move-Item "alex.exe" "$installDir\alex.exe"

# 添加到 PATH (重启 PowerShell 后生效)
$path = [Environment]::GetEnvironmentVariable("Path", "User")
[Environment]::SetEnvironmentVariable("Path", "$path;$installDir", "User")
```

## ⚙️ 高级安装选项

### 安装脚本参数

#### Linux/macOS 脚本参数

```bash
./install.sh --help
```

可用选项：
- `--version VERSION`: 安装指定版本
- `--repo REPO`: 指定 GitHub 仓库
- `--install-dir DIR`: 指定安装目录

示例：
```bash
# 安装特定版本
./install.sh --version v1.0.0

# 安装到自定义目录
./install.sh --install-dir /opt/alex

# 从不同仓库安装
./install.sh --repo your-org/your-repo
```

#### Windows 脚本参数

```powershell
.\install.ps1 -Help
```

可用参数：
- `-Version VERSION`: 安装指定版本
- `-Repository REPO`: 指定 GitHub 仓库  
- `-InstallDir DIR`: 指定安装目录

示例：
```powershell
# 安装特定版本
.\install.ps1 -Version v1.0.0

# 安装到自定义目录
.\install.ps1 -InstallDir "C:\Program Files\Alex"

# 从不同仓库安装
.\install.ps1 -Repository "your-org/your-repo"
```

### 验证安装

安装完成后，验证是否正确安装：

```bash
# 查看版本
alex --version

# 查看帮助
alex --help

# 运行简单命令
alex "What tools are available?"
```

## 🔧 构建配置

### 支持的平台

| 操作系统 | 架构 | 状态 |
|----------|------|------|
| Linux | AMD64 | ✅ |
| Linux | ARM64 | ✅ |
| macOS | AMD64 (Intel) | ✅ |
| macOS | ARM64 (Apple Silicon) | ✅ |
| Windows | AMD64 | ✅ |

### GitHub Actions 自动构建

项目使用 GitHub Actions 自动构建和发布：

- **触发条件**: 推送 tag (格式: `v*.*.*`) 或手动触发
- **构建矩阵**: 支持所有主流平台和架构
- **产物**: 生成跨平台二进制文件并自动发布到 Releases
- **校验**: 自动生成 SHA256 校验文件

### 本地构建

如果你想从源码构建：

```bash
# 克隆仓库
git clone https://github.com/ckl/Alex-Code.git
cd Alex-Code

# 安装依赖
make deps

# 构建当前平台
make build

# 构建所有平台
make build-all

# 构建结果在 build/ 目录下
ls build/
```

## 🛠️ 故障排除

### 常见问题

1. **权限错误 (Linux/macOS)**
   ```bash
   # 如果安装到系统目录需要 sudo
   sudo ./install.sh
   
   # 或者安装到用户目录
   ./install.sh --install-dir ~/.local/bin
   ```

2. **PowerShell 执行策略错误 (Windows)**
   ```powershell
   # 设置执行策略
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   
   # 或者绕过执行策略运行
   powershell -ExecutionPolicy Bypass -File install.ps1
   ```

3. **PATH 环境变量问题**
   
   安装后如果无法找到 `alex` 命令，请：
   
   - **Linux/macOS**: 将安装目录添加到 `~/.bashrc` 或 `~/.zshrc`
     ```bash
     echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc
     source ~/.bashrc
     ```
   
   - **Windows**: 重启 PowerShell 或添加到系统 PATH

4. **网络连接问题**
   
   如果下载失败，可以：
   - 使用代理: `export https_proxy=http://your-proxy:port`
   - 手动下载二进制文件后本地安装

### 卸载

#### Linux/macOS
```bash
# 删除二进制文件
sudo rm /usr/local/bin/alex
# 或者
rm ~/.local/bin/alex

# 清理配置文件（可选）
rm -rf ~/.alex
```

#### Windows
```powershell
# 删除安装目录
Remove-Item "$env:LOCALAPPDATA\Alex" -Recurse -Force

# 从 PATH 中移除（手动编辑环境变量）
```

## 📚 更多信息

- [使用指南](quickstart.md)
- [API 参考](../reference/api-reference.md)
- [开发文档](../architecture/01-architecture-overview.md)
- [问题反馈](https://github.com/ckl/Alex-Code/issues)

## 🤝 贡献

欢迎贡献安装脚本的改进和新平台支持！请查看 [贡献指南](../README.md#contributing) 了解更多信息。 