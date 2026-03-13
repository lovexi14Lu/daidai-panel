# 呆呆面板 GitHub 上传与发布指南

## 一、首次上传到 GitHub

### 1. 创建 GitHub 仓库

在 GitHub 上创建一个新仓库：`daidai-panel`（不要勾选初始化 README）

### 2. 配置 Git 并推送

```bash
cd /d/爱学习的呆子/呆呆面板开发

# 添加远程仓库
git remote add origin https://github.com/linzixuanzz/daidai-panel.git

# 添加所有需要的文件到暂存区
git add .

# 提交
git commit -m "Initial commit: 呆呆面板 v1.0.0"

# 推送到 GitHub
git push -u origin main
```

### 3. 确认 .gitignore

推送前确保 `.gitignore` 包含以下内容，避免上传不必要的文件：

```
node_modules/
web/dist/
*.exe
__pycache__/
.env
*.db
data/
```

## 二、创建 Release（手动方式）

### 1. 打 Tag

```bash
git tag v0.2.0
git push origin v0.2.0
```

### 2. 用 GitHub CLI 创建 Release

先安装 GitHub CLI：https://cli.github.com/

```bash
# 登录（首次使用）
gh auth login

# 创建 Release 并附带说明
gh release create v0.2.0 \
  --title "v0.2.0" \
  --notes "## 呆呆面板 v0.2.0

### 功能特性
- 定时任务管理（Cron 调度、重试、超时、前后置钩子）
- 脚本管理（在线代码编辑器）
- 环境变量管理（分组、拖拽排序、批量导入导出）
- 订阅管理（Git 仓库自动拉取同步）
- 依赖管理（pip / npm）
- 18 种消息推送渠道
- Open API（App Key / App Secret 认证）
- 系统监控（CPU / 内存 / 磁盘）
- Docker 一键部署

### 部署方式

**Docker Compose（推荐）：**
\`\`\`yaml
services:
  daidai-panel:
    image: linzixuanzz/daidai-panel:latest
    container_name: daidai-panel
    restart: unless-stopped
    ports:
      - \"5700:5700\"
    volumes:
      - ./data:/app/data
    environment:
      - TZ=Asia/Shanghai
\`\`\`
"
```

### 3. 上传编译好的二进制文件到 Release

```bash
# 编译 linux amd64
cd server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o daidai-linux-amd64 .

# 编译 linux arm64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o daidai-linux-arm64 .

# 打包（包含前端和配置文件）
cd ..
mkdir -p release
cp server/daidai-linux-amd64 release/
cp server/config.yaml release/
cp -r web/dist release/web
cd release
tar -czvf daidai-linux-amd64.tar.gz daidai-linux-amd64 config.yaml web/
cd ..

# 上传到已有的 Release
gh release upload v0.2.0 release/daidai-linux-amd64.tar.gz
```

## 三、自动发布（已配置 GitHub Actions）

项目已配置 `.github/workflows/release.yml`，只需要打 Tag 推送即可自动编译发布：

```bash
# 打 Tag
git tag v0.2.0

# 推送 Tag，GitHub Actions 会自动编译并创建 Release
git push origin v0.2.0
```

Actions 会自动：
1. 编译前端（npm ci + build）
2. 交叉编译 Go 后端（linux/amd64 + linux/arm64）
3. 打包成 tar.gz
4. 创建 Release 并上传附件

## 四、后续更新流程

```bash
# 1. 修改代码后提交
git add .
git commit -m "feat: 新增xxx功能"
git push origin main

# 2. 需要发新版本时打 Tag
git tag v0.2.1
git push origin v0.2.1
# GitHub Actions 自动处理剩下的
```

## 五、常用 GitHub CLI 命令

```bash
# 查看所有 Release
gh release list

# 查看某个 Release 详情
gh release view v0.2.0

# 删除某个 Release
gh release delete v0.2.0

# 编辑 Release 说明
gh release edit v0.2.0 --notes "新的说明内容"

# 给已有 Release 追加文件
gh release upload v0.2.0 新文件.tar.gz

# 删除 Release 中的某个文件
gh release delete-asset v0.2.0 旧文件.tar.gz
```
