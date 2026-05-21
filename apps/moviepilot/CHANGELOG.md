## 2026-05-20

- 切换为 host 网络模式
- 端口调整：Web UI 4000，API 4001
- UMASK 改为 000
- 管理员用户名可自定义（默认 admin）
- 移除 GITHUB_PROXY 环境变量
- 修复 `cmd/main` 状态检测 exit 1 → return 1

## 2026-05-16

- 安装时可自定义 admin 初始密码（默认 password）
- 简化安装流程

## 2026-02-14

- 首次发布
- 基于 Docker 容器部署
