# AxisRelay 生产部署

## 上线前

1. 从 `.env.example` 复制 `.env`，替换数据库、Redis、后台密钥和 SMTP 配置。
2. 将 `AXISRELAY_PUBLIC_BASE_URL` 设置为用户实际访问的 HTTPS 根地址。
3. 开放注册时，必须配置 SMTP 主机、发件地址和 TLS 模式；未配置时注册接口会明确失败。
4. 反向代理应把 `/`、`/auth/*`、`/console/*` 交给 Web 应用，把 `/admin/*` 交给后台应用，并原样转发 `/api/*`、`/v1/*` 和 WebSocket 升级头。

## 启动与升级

```bash
docker compose pull
docker compose up -d
docker compose ps
```

应用启动时按顺序执行数据库迁移。升级前先备份数据库与图片资产；升级完成后检查 `/health`，再依次完成注册、邮箱验证、登录、密码找回和 API 请求冒烟。

## 备份

```bash
docker exec axisrelay-mysql mysqldump -uaxisrelay -p axisrelay > axisrelay-backup.sql
docker run --rm -v axisrelay_image-assets:/data -v "$PWD":/backup alpine tar czf /backup/axisrelay-images.tgz -C /data .
```

不要把 `.env`、数据库备份或真实 SMTP 凭据加入源码仓库和发布压缩包。

## 前端边界

- `frontend/web`：官网、认证、用户中心和公开自助门户，构建到 `frontend/dist/web`。
- `frontend/admin`：后台管理系统，构建到 `frontend/dist/admin`。
- 官网和用户中心不展示后台入口；后台仅可通过已知的 `/admin` 地址直接访问。
