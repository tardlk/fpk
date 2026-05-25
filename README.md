# fpk

飞牛 fnOS 第三方应用打包仓库，自动跟踪上游版本，构建 `.fpk` 安装包。

## 应用

| App | 端口 | 类型 | 说明 |
|-----|:---:|:---:|------|
| **Plex** | 32400 | Native | 媒体服务器，支持硬件转码 |
| **Emby** | 8096 | Native | 媒体管理与流式传输 |
| **Jellyfin** | 8097 | Native | 开源媒体系统，内置 FFmpeg |
| **qBittorrent** | 8085 | Native | BT 客户端，默认 admin/adminadmin |
| **OpenList** | 5244 | Native | 文件列表 / WebDAV 服务器 |
| **Alist** | 5246 | Native | 网盘聚合，挂载阿里云盘、百度网盘等 |
| **FNet** | N/A | Native | 网络优化，BBR 拥塞控制 + hosts 编辑 |
| **Firefox** | 5801 | Docker | 远程浏览器，中文支持，可设 VNC 密码 |
| **MoviePilot** | 3000 | Docker | 影视自动化管理，默认 admin/password |

## 安装

从 [Releases](https://github.com/tardlk/fpk/releases) 下载 `.fpk`，在 fnOS 应用中心手动安装。

## 本地构建

```bash
# Native 应用 — 以 plex 为例
cd apps/plex && ./update_plex.sh            # 最新版，自动检测架构
cd apps/plex && ./update_plex.sh --arch arm  # 指定 ARM

# Docker 应用 — CI 自动构建
```

## 新增应用

```bash
./scripts/new-app.sh <name> "<display>" <port>
```

详见 [新增应用指南](docs/adding-new-apps.md)。

## 说明

基于 [conversun/fnos-apps](https://github.com/conversun/fnos-apps) 修改，仅挑选部分自用应用并调整配置。感谢 conversun 及所有贡献者的工作。
