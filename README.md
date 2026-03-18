<p align="center">
  <img src="./images/图标.png" alt="呆呆面板" width="120">
</p>

<h1 align="center">呆呆面板</h1>

<p align="center">
  <em>轻量、现代的定时任务管理面板，Docker 一键部署，开箱即用</em>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Vue-3-4FC08D?logo=vue.js&logoColor=white" alt="Vue3">
  <img src="https://img.shields.io/badge/Element%20Plus-2.x-409EFF?logo=element&logoColor=white" alt="Element Plus">
  <img src="https://img.shields.io/badge/SQLite-3-003B57?logo=sqlite&logoColor=white" alt="SQLite">
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white" alt="Docker">
</p>

---

呆呆面板 (Daidai Panel) 是一款轻量级定时任务管理平台，采用 Go (Gin) + Vue3 (Element Plus) + SQLite 架构，专注于脚本托管与自动化任务调度。支持 Python、Node.js、Shell 等多语言脚本的定时执行与可视化管理，内置 18 种消息推送渠道、订阅管理、环境变量、依赖管理、Open API 等功能。Docker 一键部署，开箱即用。

## 功能特性

- **定时任务** — Cron 表达式调度，支持重试、超时、任务依赖、前后置钩子
- **脚本管理** — 在线代码编辑器，支持 Python、Node.js、Shell、TypeScript，拖拽移动文件
- **执行日志** — SSE 实时日志流，历史日志查看与自动清理
- **环境变量** — 分组管理、拖拽排序、批量导入导出（兼容青龙格式）
- **订阅管理** — 自动从 Git 仓库拉取脚本，支持定期同步
- **依赖管理** — 可视化安装/卸载 Python (pip) 和 Node.js (npm) 依赖
- **通知推送** — Bark、Telegram、Server酱、企业微信、钉钉、飞书等 18 种渠道
- **开放 API** — App Key / App Secret 认证，支持第三方系统对接
- **系统安全** — 双因素认证 (2FA)、IP 白名单、登录日志、会话管理
- **数据备份** — 一键备份与恢复，导出全部数据
- **系统监控** — 实时 CPU / 内存 / 磁盘监控，任务执行趋势统计

<details>
<summary><b>点击展开查看详细功能</b></summary>

### 定时任务管理
- 标准 Cron 表达式调度
- 常用时间规则快捷选择
- 任务启用/禁用状态切换
- 手动触发执行
- 任务超时控制与重试机制
- 前后置钩子（任务依赖链）
- 多实例并发控制

### 脚本文件管理
- 在线代码编辑器（语法高亮）
- 支持创建、重命名、删除文件
- 支持文件上传与拖拽移动
- 脚本版本管理
- 调试运行与实时日志输出

### 执行日志
- SSE 实时日志流
- 执行状态追踪（成功/失败/超时/手动终止）
- 执行耗时统计
- 日志自动清理策略

### 环境变量
- 安全存储敏感配置
- 变量值脱敏显示
- 分组管理与拖拽排序
- 批量导入导出（兼容青龙格式）
- 任务执行时自动注入

### 订阅管理
- Git 仓库自动拉取
- 定期同步（Cron 调度）
- SSH Key / Token 认证
- 白名单/黑名单过滤

### 消息推送
- 18 种主流推送渠道
- 任务执行结果通知
- 系统事件告警
- 自定义推送模板

### 系统设置
- 双因素认证 (2FA / TOTP)
- IP 白名单
- 登录日志与会话管理
- 数据备份与恢复
- 面板标题与图标自定义

</details>

## 效果图

<details>
<summary><b>点击展开查看界面截图</b></summary>

| 功能 | 截图 |
|------|------|
| 登录页面 | ![登录](./images/登录.png) |
| 仪表盘 | ![仪表盘](./images/仪表盘.png) |
| 定时任务 | ![定时任务](./images/定时任务.png) |
| 脚本管理 | ![脚本管理](./images/脚本管理.png) |
| 环境变量 | ![环境变量](./images/环境变量.png) |
| 订阅管理 | ![订阅管理](./images/订阅管理.png) |
| 消息通知 | ![消息通知](./images/消息通知.png) |
| 依赖管理 | ![依赖管理](./images/依赖管理.png) |
| API 文档 | ![API文档](./images/接口文档.png) |

</details>

## 快速部署

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
      - ./Dumb-Panel:/app/Dumb-Panel
      - /var/run/docker.sock:/var/run/docker.sock  
    environment:
      - TZ=Asia/Shanghai
      - CONTAINER_NAME=daidai-panel
      - IMAGE_NAME=linzixuanzz/daidai-panel:latest
```

```bash
docker compose up -d
```

### Docker Run

```bash
docker run -d \
  --pull=always \
  --name daidai-panel \
  --restart unless-stopped \
  -p 5700:5700 \
  -v $(pwd)/Dumb-Panel:/app/Dumb-Panel \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e TZ=Asia/Shanghai \
  -e CONTAINER_NAME=daidai-panel \
  -e IMAGE_NAME=linzixuanzz/daidai-panel:latest \
  linzixuanzz/daidai-panel:latest
```

启动后访问：`http://localhost:5700`

首次使用需要初始化管理员账号。

> **说明**：挂载 `/var/run/docker.sock` 是为了支持面板内一键更新功能。如果不需要此功能，可以移除该挂载。

### 本地开发运行

#### 环境要求

- Go 1.25+
- Node.js 18+（推荐 20+）
- npm 或 pnpm

#### 启动后端

```bash
cd server
go run .
```

后端默认监听 `5701` 端口，读取同目录下的 `config.yaml` 作为配置文件。

#### 启动前端

```bash
cd web
npm install
npm run dev
```

前端 Vite 开发服务器默认运行在 `5173` 端口，已配置将 `/api` 请求代理到 `http://localhost:5701`。

启动后访问：`http://localhost:5173`

#### 本地端口修改

**修改后端端口**：编辑 `server/config.yaml`

```yaml
server:
  port: 5701    # 改为你想要的端口
```

修改后端端口后，需要同步修改前端代理地址。编辑 `web/vite.config.ts`：

```typescript
server: {
  port: 5173,   // 前端端口，按需修改
  proxy: {
    '/api': {
      target: 'http://localhost:5701',  // 改为对应的后端端口
      changeOrigin: true
    }
  }
}
```

#### 构建生产版本

```bash
# 构建前端
cd web
npm run build

# 构建后端
cd server
go build -o daidai-panel .
```

前端构建产物在 `web/dist/`，需配合 Nginx 或其他静态服务器部署，并反向代理 `/api` 到后端。

### 自定义端口（Docker）

Docker 部署时默认面板端口为 5700。如需修改宿主机访问端口，只需更改 `-p` 左侧的端口号即可：

```bash
# 示例：通过宿主机 8080 端口访问面板
docker run -d \
  --pull=always \
  --name daidai-panel \
  --restart unless-stopped \
  -p 8080:5700 \
  -v $(pwd)/Dumb-Panel:/app/Dumb-Panel \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e TZ=Asia/Shanghai \
  linzixuanzz/daidai-panel:latest
```

如果需要同时修改容器内部端口，通过 `PANEL_PORT` 环境变量指定，并保持 `-p` 右侧端口与其一致：

```bash
# 示例：容器内部使用 7100 端口，宿主机通过 8080 访问
docker run -d \
  --pull=always \
  --name daidai-panel \
  --restart unless-stopped \
  -p 8080:7100 \
  -e PANEL_PORT=7100 \
  -v $(pwd)/Dumb-Panel:/app/Dumb-Panel \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e TZ=Asia/Shanghai \
  linzixuanzz/daidai-panel:latest
```

> **注意**：`-p` 右侧的容器端口必须与 `PANEL_PORT` 一致，否则面板将无法访问。

## 多架构支持

镜像同时支持 `linux/amd64` 和 `linux/arm64`，可在 x86 服务器和 ARM 设备（如树莓派、Oracle ARM 云服务器）上直接运行。

## 更新方法

### 方式一：面板内一键更新（推荐）

进入「系统设置」→「概览」，点击「检查系统更新」，如有新版本会提示一键更新。

### 方式二：手动更新

```bash
docker pull linzixuanzz/daidai-panel:latest
docker compose up -d
```

## 数据目录

```
./Dumb-Panel/
├── daidai.db          # SQLite 数据库
├── scripts/           # 脚本文件存储
├── logs/              # 执行日志
└── backups/           # 数据备份
```

## 技术栈

| 层 | 技术 |
|----|------|
| 前端 | Vue 3 + TypeScript + Element Plus + Pinia + Vite |
| 后端 | Go (Gin) + GORM + SQLite |
| 部署 | Nginx + Go Binary，Docker 单镜像（AMD64 / ARM64） |

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `TZ` | 时区 | `Asia/Shanghai` |
| `DATA_DIR` | 数据存储目录 | `/app/Dumb-Panel` |
| `DB_PATH` | 数据库路径 | `${DATA_DIR}/daidai.db` |
| `SERVER_PORT` | Go 服务端口 | `5701` |
| `PANEL_PORT` | 面板访问端口（容器内 Nginx 监听端口） | `5700` |

<details>
<summary><b>config.yaml 完整配置说明</b></summary>

本地开发时后端读取 `server/config.yaml`，Docker 部署时由 `entrypoint.sh` 自动生成。

```yaml
server:
  port: 5701          # 后端 API 端口
  mode: release       # debug / release

database:
  path: ./data/daidai.db    # SQLite 数据库路径

jwt:
  secret: ""                # 留空则自动生成并持久化
  access_token_expire: 480h
  refresh_token_expire: 1440h

data:
  dir: ./data               # 数据根目录
  scripts_dir: ./data/scripts
  log_dir: ./data/logs

cors:
  origins:                  # 允许的跨域来源
    - http://localhost:5173
    - http://localhost:5700
```

</details>

<details>
<summary><b>Nginx 反向代理配置（HTTPS）</b></summary>

```nginx
map $http_upgrade $connection_upgrade {
    default upgrade;
    '' close;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate     /path/to/fullchain.pem;
    ssl_certificate_key /path/to/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:5700;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;

        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;

        proxy_buffering off;
        proxy_read_timeout 300s;
    }
}
```

</details>

## 致谢

本项目的开发离不开以下优秀的开源项目：

- **[白虎面板 (Baihu Panel)](https://github.com/engigu/baihu-panel)** — 后端框架架构参考，部分代码基于白虎面板改进
- **[青龙面板 (Qinglong)](https://github.com/whyour/qinglong)** — 功能设计参考，定时任务管理、环境变量、订阅管理等核心功能借鉴自青龙面板

感谢以上项目作者的贡献！

## LICENSE

Copyright © 2026, [linzixuanzz](https://github.com/linzixuanzz). Released under the [MIT](LICENSE).
