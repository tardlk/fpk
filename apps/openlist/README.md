# OpenList for fnOS

每日自动同步 [OpenList 官方](https://github.com/OpenListTeam/OpenList) 最新版本并构建 `.fpk` 安装包。

## 下载

从 [Releases](https://github.com/tardlk/fpk/releases?q=openlist) 下载最新的 `.fpk` 文件。

## 安装

1. 根据设备架构下载对应的 `.fpk` 文件
2. fnOS 应用管理 → 手动安装 → 上传

**访问地址**: `http://<NAS-IP>:5244`

## 说明

- AList 社区分支，支持多种存储后端的文件列表程序
- 支持阿里云盘、OneDrive、百度网盘等多种存储
- 提供 WebDAV 服务器功能
- 支持文件在线预览和视频播放
- 数据存储在应用数据目录中

## 本地构建

```bash
./update_openlist.sh                        # 最新版本，自动检测架构
./update_openlist.sh --arch arm             # 指定架构
./update_openlist.sh --help                 # 查看帮助
```

## 版本标签

- `openlist/vX.Y.Z` — 首次发布
- `openlist/vX.Y.Z-r2` — 同版本打包修订

## Credits

- [OpenList](https://github.com/OpenListTeam/OpenList)
