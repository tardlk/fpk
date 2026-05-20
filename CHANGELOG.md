# Changelog

## 2026-05-20

- 修复 Emby 包装脚本缺少重启循环，退出码 3 时无法自动重启
- 修复 Docker 应用状态检测：`daemon_status` 中 `exit 1` → `return 1`
- 修复 `check_docker` 路径，对齐 fnOS 实际目录结构
- qBittorrent 日志名统一为 `${TRIM_APPNAME}-service.log`
- 为 Emby / Plex / Jellyfin 添加安装向导和卸载向导
- 为 OpenList / Firefox 添加卸载向导
- Firefox / MoviePilot 安装向导新增 Docker 镜像加速选项，自动规范用户输入
- 删除 Plex manifest 中无用的 `beta = no` 字段

## 2026-05-16

- 项目初始化，从 fnos-apps 精选 7 个应用
- Native: Plex, Emby, Jellyfin, qBittorrent, OpenList
- Docker: Firefox, MoviePilot
- 修复 Firefox VNC 密码不生效（.vncpass_clear 文件方式）
- 修复 MoviePilot SUPERUSER_PASSWORD 不生效（compose 变量替换）
