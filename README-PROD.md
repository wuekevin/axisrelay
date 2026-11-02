# AxisRelay 宝塔原生部署（不使用 Docker）

## 1. 宝塔准备

安装并启用 Nginx、MySQL 8.0 和 Redis。创建专用数据库及最小权限账号，不要使用 MySQL `root` 运行应用：

```sql
CREATE DATABASE axisrelay CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE USER 'axisrelay'@'127.0.0.1' IDENTIFIED WITH caching_sha2_password BY '替换为强密码';
GRANT ALL PRIVILEGES ON axisrelay.* TO 'axisrelay'@'127.0.0.1';
FLUSH PRIVILEGES;
```

将生产包上传到例如 `/www/wwwroot/axisrelay`，按服务器架构选择二进制：

- Intel/AMD 64 位：`bin/axisrelay-linux-amd64`
- ARM 64 位：`bin/axisrelay-linux-arm64`

```bash
cd /www/wwwroot/axisrelay
chmod 750 bin/axisrelay-linux-amd64
mkdir -p data/images logs
chmod 750 data data/images logs
```

首次运行不要复制 `.env.example` 为 `.env`。程序检测不到有效配置时会在服务端口打开 `/install` 安装页。

## 2. 首次安装

先以前台方式启动一次，确认安装地址：

```bash
cd /www/wwwroot/axisrelay
AXISRELAY_PORT=8081 ./bin/axisrelay-linux-amd64
```

若浏览器通过公网访问宝塔网站，首次启动时临时放行你当前的固定公网 IP（安装完成后从守护进程环境中删除）：

```bash
AXISRELAY_PORT=8081 AXISRELAY_INSTALL_ALLOWED_CIDR=203.0.113.7/32 ./bin/axisrelay-linux-amd64
```

在宝塔新建网站并设置反向代理到 `http://127.0.0.1:8081`，然后访问 `https://你的域名/install`。填写：

- 站点公开 HTTPS 地址；
- MySQL 主机、端口、账号、密码和数据库名；
- Redis 地址、账号、密码和 DB；
- 管理员账号和至少 12 位的强密码。

安装器会真实连接数据库与 Redis、执行数据库迁移、创建管理员账号，然后原子写入权限为 `0600` 的 `.env` 和 `data/install.lock`。密码和完整 DSN 不会写入日志。检测到已有管理员时安装会停止，不会覆盖账号或删除数据。

安装完成后，`/` 是官网，`/auth/*` 是注册登录，`/console/*` 是用户中心；后台地址为 `/admin`，使用管理员账号密码登录。官网和用户中心不展示后台入口。

## 3. 宝塔进程守护

可使用宝塔「进程守护管理器」：

- 名称：`axisrelay`
- 运行目录：`/www/wwwroot/axisrelay`
- 启动命令：`/www/wwwroot/axisrelay/bin/axisrelay-linux-amd64`
- 运行用户：建议为站点专用的低权限用户
- 自动启动：开启

不要在启动命令中放数据库或 Redis 密码。安装后的配置只保存在 `/www/wwwroot/axisrelay/.env`。

## 4. Nginx 反向代理

宝塔反向代理目标填写 `http://127.0.0.1:8081`，并保留以下请求头：

```nginx
proxy_set_header Host $host;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto $scheme;
proxy_http_version 1.1;
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "upgrade";
proxy_read_timeout 600s;
proxy_buffering off;
```

在宝塔申请并强制 HTTPS。防火墙不需要对公网开放 8081，只开放 80/443。

## 5. SMTP 与注册

开放用户注册前，在 `.env` 中补齐 `AXISRELAY_PUBLIC_BASE_URL`、SMTP 主机、端口、发件地址和 TLS 配置；缺少生产邮件配置时，注册流程会明确失败，不会静默跳过邮箱验证。修改 `.env` 后重启守护进程。

## 6. 升级与备份

升级前备份数据库、`.env`、`data/install.lock` 和图片目录：

```bash
mysqldump -h127.0.0.1 -uaxisrelay -p --single-transaction axisrelay > axisrelay-backup.sql
tar czf axisrelay-files-backup.tgz .env data/install.lock data/images
```

停止守护进程，替换源码和二进制，但保留上述配置与数据，再启动。应用启动时只执行向前兼容迁移，不会重置管理员账号。

升级自旧版管理密钥时：若数据库还没有管理员账号，旧 `AXISRELAY_ADMIN_SECRET` 会一次性转换为用户名 `admin` 的初始密码；首次登录后必须改成新密码。转换完成后不再接受 `X-Admin-Key`。

## 7. 上线检查

```bash
curl -fsS https://你的域名/health
```

随后人工检查官网、用户注册登录、用户中心、`/admin` 账号密码登录、创建 API Key 和一次真实 API 请求。不要把 `.env`、数据库备份或真实凭据加入源码仓库或发布压缩包。

前端源码边界：`frontend/web` 是官网和用户端，`frontend/admin` 是后台；构建产物分别在 `frontend/dist/web` 与 `frontend/dist/admin`。
