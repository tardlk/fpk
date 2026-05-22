# PROJECT KNOWLEDGE BASE

## OVERVIEW

精选 fnOS 第三方应用打包仓库，8 个应用。Pure bash — 下载上游二进制，合并共享生命周期框架，输出 `.fpk`。CI 每日自动同步上游版本。

## STRUCTURE

```
fpk/
├── shared/              # 生命周期框架（所有应用共用）
│   ├── cmd/             # 守护进程管理、安装/升级/卸载钩子
│   └── wizard/          # 默认卸载向导（各应用有独立 wizard 时优先使用）
├── apps/
│   ├── plex/            # Plex 媒体服务器 (32400, .deb 提取)
│   ├── emby/            # Emby 媒体服务器 (8096, .deb 提取)
│   ├── jellyfin/        # Jellyfin 开源媒体系统 (8097, .deb 提取)
│   ├── qbittorrent/     # qBittorrent BT 客户端 (8085, 静态二进制)
│   ├── openlist/        # OpenList 文件列表 (5244, Go 单二进制)
│   ├── fnet/            # FNet 网络优化 (Unix socket, Go)
│   ├── firefox/         # Firefox 远程浏览器 (5801, Docker)
│   └── moviepilot/      # MoviePilot 影视管理 (3000, Docker)
├── scripts/
│   ├── build-fpk.sh     # 通用 fpk 打包器
│   ├── new-app.sh       # 脚手架
│   ├── apps/            # 每个应用的构建契约
│   ├── lib/             # 共享构建工具 update-common.sh
│   ├── ci/              # CI 辅助脚本
│   └── test/            # 测试脚本
└── .github/workflows/   # CI/CD
```

## WHERE TO LOOK

| Task | Location |
|------|----------|
| 应用生命周期 | `shared/cmd/common` |
| 应用特定服务配置 | `apps/*/fnos/cmd/service-setup` |
| 构建契约 | `scripts/apps/<app>/` (meta.env, build.sh, get-latest-version.sh, release-notes.tpl) |
| 通用打包器 | `scripts/build-fpk.sh` |
| CI 入口 | `.github/workflows/build-apps.yml` |

## UNIQUE STYLES

- **Plex / Emby / Jellyfin**: .deb 提取，需要 `video` + `render` 组硬件转码
- **qBittorrent**: 预配置 admin/adminadmin，中文 locale，关闭 CSRF 适配 fnOS 反向代理
- **OpenList**: Go 单二进制，最简应用
- **Firefox**: Docker，VNC 密码通过 wizard 设置
- **MoviePilot**: Docker v2，${wizard_password} compose 变量替换
- **FNet**: Go 编译，Unix socket 通信，BBR 开关 + hosts 编辑

## CONVENTIONS

- 100% bash
- Manifest 对齐第 16 列
- fpk = tar.gz: app.tgz + cmd/ + config/ + wizard/ + manifest + icons
- 双架构：x86 + arm
- 不修改上游二进制，仅下载重打包
