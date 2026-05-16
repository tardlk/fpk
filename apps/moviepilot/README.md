# MoviePilot for fnOS

每日自动同步 [MoviePilot 官方](https://github.com/jxxghp/MoviePilot) 最新版本并构建 `.fpk` 安装包。

## 下载

从 [Releases](https://github.com/tardlk/fpk/releases?q=moviepilot) 下载最新的 `.fpk` 文件。

## 安装

1. 根据设备架构下载对应的 `.fpk` 文件
2. fnOS 应用管理 → 手动安装 → 上传

**访问地址**: `http://<NAS-IP>:3000`

## 说明

- 电影/电视剧自动化管理工具
- 支持自动订阅、下载、整理和通知
- 基于 Docker 部署，首次启动需要拉取镜像
- 内置豆瓣、TMDB 等数据源对接
- 基于 Docker 部署，首次启动需要拉取镜像

## 本地构建

```bash
cd apps/moviepilot && bash ../../scripts/build-fpk.sh . app.tgz
```

## Credits

- [MoviePilot](https://github.com/jxxghp/MoviePilot)
