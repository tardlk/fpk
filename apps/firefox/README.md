# Firefox for fnOS

每日自动同步 [Firefox 官方](https://github.com/jlesage/docker-firefox) 最新版本并构建 `.fpk` 安装包。

## 下载

从 [Releases](https://github.com/tardlk/fpk/releases?q=firefox) 下载最新的 `.fpk` 文件。

## 安装

1. 根据设备架构下载对应的 `.fpk` 文件
2. fnOS 应用管理 → 手动安装 → 上传

**访问地址**: `http://<NAS-IP>:5801`

## 说明

- 远程浏览器，支持中文界面和输入
- 无需本地安装，通过浏览器即可访问
- 支持书签同步和扩展插件
- 适合在无桌面环境的 NAS 上使用
- 基于 Docker 部署，首次启动需要拉取镜像

## 本地构建

```bash
cd apps/firefox && bash ../../scripts/build-fpk.sh . app.tgz
```

## Credits

- [Firefox](https://github.com/jlesage/docker-firefox)
