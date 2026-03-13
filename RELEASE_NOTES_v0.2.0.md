# 🎉 呆呆面板 v1.0.0 — 架构重构正式版

> 本版本为架构重构后的首个正式版本，从 Python Flask + React 全面迁移至 Go + Vue3，性能与体验大幅提升。

## ⚠️ 重大变更

- **后端重写**：从 Python (Flask + SQLAlchemy + SQLite) 迁移至 **Go (Gin + GORM + PostgreSQL)**，启动速度和资源占用显著优化
- **前端重写**：从 React 18 + Ant Design 迁移至 **Vue 3 + TypeScript + Element Plus + Pinia**
- **部署架构**：统一为 Nginx + Go Binary 的 Docker 单镜像方案，端口 `5700`
- **数据不兼容**：由于底层架构完全重写，v0.1.x 的数据无法直接迁移，需重新初始化

## ✨ 新增功能

- 全新登录页面，支持互动角色动画
- 页面切换动画优化（纯 opacity 过渡，无闪屏卡顿）
- 动态页面标题（呆呆面板 - 仪表盘 / 环境变量 / 定时任务 等）
- API 文档页面内置，支持在线查看所有接口说明
- Open API 认证体系（App Key / App Secret），支持第三方插件对接
- GitHub Actions CI/CD 自动编译发布（推送 Tag 自动构建 linux/amd64 + linux/arm64）
- 系统设置页面支持在线检查版本更新

## 🐛 修复

- 修复 API 文档页面空白问题
- 修复第三方插件对接文档使用面板登录凭据的问题，改为 Open API 认证
- 修复 pip3 安装依赖时的 root 用户 WARNING
- 修复系统设置页面 GitHub 链接地址错误
- 修复检查更新功能 GitHub API 地址未配置的问题
- 修复设置页面概览图标未使用自定义图标的问题
- 修复 favicon 路径引用错误

## 🎨 界面优化

- 更新自定义 favicon（图标.png）
- 优化页面过渡动画，消除卡顿感
- 统一各页面标题格式
- 清理冗余的前端构建产物（从 676 个文件 108MB 优化至 124 个文件 17MB）

## 📦 部署

### Docker Compose（推荐）

```yaml
services:
  daidai-panel:
    image: linzixuanzz/daidai-panel:latest
    container_name: daidai-panel
    restart: unless-stopped
    ports:
      - "5700:5700"
    volumes:
      - ./data:/app/data
    environment:
      - TZ=Asia/Shanghai
```

```bash
docker compose up -d
```

### Docker Run

```bash
docker run -d \
  --name daidai-panel \
  --restart unless-stopped \
  -p 5700:5700 \
  -v $(pwd)/data:/app/data \
  -e TZ=Asia/Shanghai \
  linzixuanzz/daidai-panel:latest
```

访问：`http://localhost:5700`

## 📚 文档

- 完整部署文档：https://github.com/linzixuanzz/daidai-panel#快速部署
- API 文档：面板内置，访问 `/api-docs` 页面查看

## 🔗 相关链接

- GitHub 仓库：https://github.com/linzixuanzz/daidai-panel
- Docker Hub：https://hub.docker.com/r/linzixuanzz/daidai-panel
- 问题反馈：https://github.com/linzixuanzz/daidai-panel/issues

## 💡 技术栈

| 层 | 技术 |
|----|------|
| 前端 | Vue 3 + TypeScript + Element Plus + Pinia + Vite |
| 后端 | Go (Gin) + GORM + PostgreSQL |
| 部署 | Nginx + Go Binary，Docker 单镜像（AMD64 / ARM64） |

---

**首次使用需要初始化管理员账号。**
