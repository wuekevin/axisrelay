#!/usr/bin/env bash
# ============================================================
#  axisrelay 交互式部署脚本
#  用法: bash deploy.sh
# ============================================================

set -euo pipefail

# ---------- 颜色 ----------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# ---------- 工具函数 ----------
info()    { printf "${CYAN}▸ %s${NC}\n" "$*"; }
success() { printf "${GREEN}✔ %s${NC}\n" "$*"; }
warn()    { printf "${YELLOW}⚠ %s${NC}\n" "$*"; }
error()   { printf "${RED}✘ %s${NC}\n" "$*"; exit 1; }

banner() {
  printf "\n${BOLD}${CYAN}"
  cat << 'EOF'
   ___          _           ____    _    ____ ___
  / __\ ___  __| | _____  _|___ \  / \  |  _ \_ _|
 / /   / _ \/ _` |/ _ \ \/ / __) |/ _ \ | |_) | |
/ /___| (_) | (_| |  __/>  < / __// ___ \|  __/| |
\____/ \___/ \__,_|\___/_/\_\_____/_/   \_\_|  |___|

EOF
  printf "${NC}"
  echo "  交互式部署脚本 v1.1"
  echo "  ────────────────────────────────────────"
  echo ""
}

# 输入源：兼容 `bash <(curl ...)` / 管道执行场景，强制从终端读取
_INPUT_FD="/dev/tty"
if [[ ! -r "$_INPUT_FD" ]]; then
  _INPUT_FD="/dev/stdin"
fi

# 读取用户输入，支持默认值
ask() {
  local prompt="$1" default="$2" varname="$3"
  if [[ -n "$default" ]]; then
    printf "${BOLD}%s${NC} [${GREEN}%s${NC}]: " "$prompt" "$default"
  else
    printf "${BOLD}%s${NC}: " "$prompt"
  fi
  read -r input < "$_INPUT_FD"
  printf -v "$varname" "%s" "${input:-$default}"
}

# 读取密码（不回显）
ask_secret() {
  local prompt="$1" default="$2" varname="$3"
  if [[ -n "$default" ]]; then
    printf "${BOLD}%s${NC} [${GREEN}%s${NC}]: " "$prompt" "(已设置)"
  else
    printf "${BOLD}%s${NC} (留空则自动生成): " "$prompt"
  fi
  read -rs input < "$_INPUT_FD"
  echo ""
  printf -v "$varname" "%s" "${input:-$default}"
}

# 生成随机密钥
gen_secret() {
  if command -v openssl &>/dev/null; then
    openssl rand -hex 16
  elif [[ -r /dev/urandom ]]; then
    head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n' | head -c 32
  else
    date +%s%N | sha256sum | head -c 32
  fi
}

# ---------- 自举：确保位于 axisrelay 仓库目录 ----------
# 触发条件:
#   1) 通过 `bash <(curl ...)` 远程拉起 (BASH_SOURCE 不是真实文件)
#   2) 或当前目录缺少必要的 compose / deploy.sh 文件
# 行为:
#   - 若已在仓库目录: 直接返回
#   - 否则: clone 仓库到 ./axisrelay，切入并 exec ./deploy.sh
REPO_URL="${AXISRELAY_REPO_URL:-https://github.com/wuekevin/axisrelay.git}"
REPO_BRANCH="${AXISRELAY_REPO_BRANCH:-main}"
REPO_DIR_NAME="${AXISRELAY_DIR_NAME:-axisrelay}"
EXISTING_ENV_FILE=".env"

env_default() {
  local key="$1" fallback="${2:-}" value=""

  if [[ -f "$EXISTING_ENV_FILE" ]]; then
    value="$(awk -v target="$key" '
      /^[[:space:]]*($|#)/ { next }
      {
        line=$0
        sub(/^[[:space:]]*export[[:space:]]+/, "", line)
        pos=index(line, "=")
        if (pos == 0) next

        key=substr(line, 1, pos - 1)
        gsub(/^[[:space:]]+|[[:space:]]+$/, "", key)
        if (key != target) next

        value=substr(line, pos + 1)
        sub(/\r$/, "", value)
        gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)

        first=substr(value, 1, 1)
        last=substr(value, length(value), 1)
        quote=sprintf("%c", 39)
        if ((first == "\"" && last == "\"") || (first == quote && last == quote)) {
          value=substr(value, 2, length(value) - 2)
        }

        print value
        exit
      }
    ' "$EXISTING_ENV_FILE")"
  fi

  printf "%s" "${value:-$fallback}"
}

known_compose_service_exists() {
  local compose_file
  for compose_file in docker-compose.yml docker-compose.sqlite.yml docker-compose.local.yml docker-compose.sqlite.local.yml; do
    [[ -f "$compose_file" ]] || continue
    if [[ -n "$($COMPOSE_CMD -f "$compose_file" ps -q axisrelay 2>/dev/null || true)" ]]; then
      EXISTING_COMPOSE_FILE="$compose_file"
      return 0
    fi
  done
  return 1
}

detect_deployment_state() {
  DEPLOYMENT_STATE="first"
  DEPLOYMENT_REASON="未检测到 .env 或已创建的 compose 服务"
  EXISTING_COMPOSE_FILE=""

  if [[ -f "$EXISTING_ENV_FILE" ]]; then
    DEPLOYMENT_STATE="existing"
    DEPLOYMENT_REASON="检测到已有 .env"
  fi

  if known_compose_service_exists; then
    DEPLOYMENT_STATE="existing"
    if [[ -f "$EXISTING_ENV_FILE" ]]; then
      DEPLOYMENT_REASON="检测到已有 .env 和 compose 服务 ($EXISTING_COMPOSE_FILE)"
    else
      DEPLOYMENT_REASON="检测到已有 compose 服务 ($EXISTING_COMPOSE_FILE)"
    fi
  fi
}

step_deployment_route() {
  detect_deployment_state

  echo ""
  printf "${BOLD}${CYAN}━━━ 部署状态检查 ━━━${NC}\n"
  echo ""
  if [[ "$DEPLOYMENT_STATE" == "first" ]]; then
    success "检测结果: 首次部署"
    info "$DEPLOYMENT_REASON"
    success "部署线路: 完整部署向导"
    return 0
  fi

  success "检测结果: 已有部署"
  info "$DEPLOYMENT_REASON"
  if [[ -f "$EXISTING_ENV_FILE" ]]; then
    success "已有 .env 将作为交互默认值"
  fi
  success "部署线路: 完整部署向导"
}

is_axisrelay_repo() {
  [[ -f "docker-compose.yml" ]] && [[ -f "deploy.sh" ]] \
    && grep -q '^name: axisrelay' docker-compose.yml 2>/dev/null
}

bootstrap_repo() {
  # 已经在仓库目录里：什么都不做
  if is_axisrelay_repo; then
    success "检测到当前目录为 axisrelay 仓库"
    return 0
  fi

  warn "当前目录不是 axisrelay 仓库，进入自动拉取流程"

  if ! command -v git >/dev/null 2>&1; then
    error "未找到 git，请先安装 git 后重试"
  fi

  # 如果同名目录已存在且是仓库，直接复用
  if [[ -d "$REPO_DIR_NAME/.git" ]]; then
    info "发现已有目录 $REPO_DIR_NAME，尝试更新到最新代码..."
    (cd "$REPO_DIR_NAME" && git fetch --depth=1 origin "$REPO_BRANCH" && git reset --hard "origin/$REPO_BRANCH") \
      || warn "拉取更新失败，将沿用已有代码继续部署"
  elif [[ -e "$REPO_DIR_NAME" ]]; then
    error "目录 $REPO_DIR_NAME 已存在但不是 git 仓库，请手动处理后重试"
  else
    info "克隆仓库: $REPO_URL ($REPO_BRANCH) → ./$REPO_DIR_NAME"
    git clone --depth=1 --branch "$REPO_BRANCH" "$REPO_URL" "$REPO_DIR_NAME" \
      || error "git clone 失败，请检查网络或仓库地址"
  fi

  cd "$REPO_DIR_NAME" || error "无法进入 $REPO_DIR_NAME 目录"

  if ! is_axisrelay_repo; then
    error "克隆后仍未识别为 axisrelay 仓库，请手动检查"
  fi

  success "已切换到 $(pwd)"
  info "重新运行 ./deploy.sh 完成部署..."
  echo ""
  # exec 掉本进程，避免远程脚本/旧上下文继续运行
  exec bash ./deploy.sh "$@"
}

update_repo_code() {
  if [[ "${AXISRELAY_SKIP_GIT_PULL:-}" == "1" || "${AXISRELAY_SKIP_GIT_PULL:-}" == "true" ]]; then
    warn "已跳过自动拉取最新代码 (AXISRELAY_SKIP_GIT_PULL=${AXISRELAY_SKIP_GIT_PULL})"
    return 0
  fi

  if ! command -v git >/dev/null 2>&1 || ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    warn "当前目录不是 git 仓库，跳过自动拉取最新代码"
    return 0
  fi

  if ! git diff --quiet 2>/dev/null || ! git diff --cached --quiet 2>/dev/null; then
    warn "检测到本地已跟踪文件存在未提交更改，跳过自动拉取最新代码，避免覆盖本地修改"
    return 0
  fi

  local branch="${REPO_BRANCH:-main}"
  if [[ -z "$branch" ]]; then
    branch="$(git branch --show-current 2>/dev/null || true)"
  fi
  if [[ -z "$branch" ]]; then
    warn "无法识别当前分支，跳过自动拉取最新代码"
    return 0
  fi

  info "拉取最新代码: origin/$branch"
  if git fetch origin "$branch" && git pull --ff-only origin "$branch"; then
    success "代码已更新到最新可快进版本"
  else
    warn "自动拉取最新代码失败，将沿用当前代码继续部署"
  fi
}

# ---------- 前置检查 ----------
preflight() {
  info "检查运行环境..."

  if ! command -v docker &>/dev/null; then
    error "未找到 docker，请先安装 Docker"
  fi

  if docker compose version &>/dev/null; then
    COMPOSE_CMD="docker compose"
  elif command -v docker-compose &>/dev/null; then
    COMPOSE_CMD="docker-compose"
  else
    error "未找到 docker compose，请安装 Docker Compose v2+"
  fi

  success "Docker 环境就绪 ($COMPOSE_CMD)"
}

# ---------- 第一步：端口 ----------
step_port() {
  echo ""
  printf "${BOLD}${CYAN}━━━ 1/6 服务端口 ━━━${NC}\n"
  ask "服务监听端口" "$(env_default AXISRELAY_PORT "$(env_default PORT "8080")")" PORT

  if ! [[ "$PORT" =~ ^[0-9]+$ ]] || (( PORT < 1 || PORT > 65535 )); then
    error "无效端口号: $PORT"
  fi
  success "端口: $PORT"
}

# ---------- 第二步：监听范围 ----------
step_bind() {
  echo ""
  printf "${BOLD}${CYAN}━━━ 2/6 监听范围 ━━━${NC}\n"
  echo ""
  echo "  1) 仅本机访问  — 绑定 127.0.0.1，外部无法访问 (内网/反向代理后端推荐)"
  echo "  2) 全部网络    — 绑定 0.0.0.0，可通过内网/公网 IP 访问 (默认)"
  echo ""
  local bind_default bind_choice_default
  bind_default="$(env_default AXISRELAY_BIND_HOST "0.0.0.0")"
  case "$bind_default" in
    127.*|localhost)
      bind_choice_default="1"
      ;;
    *)
      bind_choice_default="2"
      ;;
  esac
  ask "请选择 (1 或 2)" "$bind_choice_default" BIND_CHOICE

  case "$BIND_CHOICE" in
    1|local|loopback|127*)
      AXISRELAY_BIND_HOST="127.0.0.1"
      BIND_MODE="loopback"
      success "监听范围: 仅本机 (127.0.0.1)"
      ;;
    2|all|public|0*)
      AXISRELAY_BIND_HOST="0.0.0.0"
      BIND_MODE="all"
      success "监听范围: 全部网络 (0.0.0.0)"
      ;;
    *)
      error "无效选择: $BIND_CHOICE"
      ;;
  esac
}

# ---------- 第三步：数据库模式 ----------
step_database() {
  echo ""
  printf "${BOLD}${CYAN}━━━ 3/6 数据库模式 ━━━${NC}\n"
  echo ""
  info "S0.3 已移除 PostgreSQL；当前部署脚本仅保留 SQLite 过渡模式，MySQL 将在后续阶段接入。"
  DB_MODE="sqlite"
  success "数据库模式: SQLite (过渡)"
  step_sqlite_config
}

step_sqlite_config() {
  echo ""
  ask "SQLite 数据文件路径 (容器内)" "$(env_default AXISRELAY_DATABASE_PATH "/data/axisrelay.db")" SQLITE_PATH
}

# ---------- 第四步：密钥 ----------
step_secrets() {
  echo ""
  printf "${BOLD}${CYAN}━━━ 4/6 安全密钥 ━━━${NC}\n"
  echo ""

  ask_secret "管理后台密钥 (AXISRELAY_ADMIN_SECRET)" "$(env_default AXISRELAY_ADMIN_SECRET "")" AXISRELAY_ADMIN_SECRET
  if [[ -z "$AXISRELAY_ADMIN_SECRET" ]]; then
    AXISRELAY_ADMIN_SECRET=$(gen_secret)
    success "已自动生成管理密钥"
  fi

  echo ""
}

# ---------- 第五步：构建方式 ----------
step_build_mode() {
  echo ""
  printf "${BOLD}${CYAN}━━━ 5/6 构建方式 ━━━${NC}\n"
  echo ""
  echo "  1) 拉取镜像 — 使用预构建镜像 (ghcr.io)，部署快"
  echo "  2) 本地构建 — 从源码编译，适合自定义修改"
  echo ""
  ask "请选择 (1 或 2)" "1" BUILD_CHOICE

  case "$BUILD_CHOICE" in
    1|pull|image)
      BUILD_MODE="image"
      success "构建方式: 拉取预构建镜像"
      ;;
    2|local|build)
      BUILD_MODE="local"
      success "构建方式: 本地源码构建"
      ;;
    *)
      error "无效选择: $BUILD_CHOICE"
      ;;
  esac
}

# ---------- 第六步：确认 ----------
step_confirm() {
  echo ""
  printf "${BOLD}${CYAN}━━━ 6/6 配置确认 ━━━${NC}\n"
  echo ""
  echo "  端口:       $PORT"
  if [[ "$BIND_MODE" == "loopback" ]]; then
    echo "  监听范围:   127.0.0.1 (仅本机访问)"
  else
    echo "  监听范围:   0.0.0.0 (全部网络)"
  fi
  echo "  数据库:     $DB_MODE"
  echo "  数据路径:   $SQLITE_PATH"
  echo "  缓存:       memory"
  echo "  构建方式:   $( [[ "$BUILD_MODE" == "image" ]] && echo "拉取镜像" || echo "本地构建" )"
  echo "  管理密钥:   ${AXISRELAY_ADMIN_SECRET}"
  echo ""
  ask "确认部署? (y/n)" "y" CONFIRM
  if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
    warn "已取消部署"
    exit 0
  fi
}

# ---------- 生成 .env ----------
generate_env() {
  info "生成 .env 文件..."

  # 备份已有 .env
  if [[ -f .env ]]; then
    cp .env ".env.bak.$(date +%Y%m%d%H%M%S)"
    warn "已备份原 .env 文件"
  fi

  cat > .env << EOF
# ============================
#  axisrelay 配置 (SQLite 模式)
#  由 deploy.sh 自动生成于 $(date '+%Y-%m-%d %H:%M:%S')
# ============================

# 服务端口
AXISRELAY_PORT=${PORT}

# 端口绑定地址 (127.0.0.1=仅本机, 0.0.0.0=全部网络)
AXISRELAY_BIND_HOST=${AXISRELAY_BIND_HOST}

# 管理后台密钥
AXISRELAY_ADMIN_SECRET=${AXISRELAY_ADMIN_SECRET}

# 数据库 — SQLite
AXISRELAY_DATABASE_DRIVER=sqlite
AXISRELAY_DATABASE_PATH=${SQLITE_PATH}

# 缓存 — 内存
AXISRELAY_CACHE_DRIVER=memory

# 时区
TZ=Asia/Shanghai
EOF


  success ".env 已生成"
}

# ---------- 选择 compose 文件 ----------
resolve_compose_file() {
  if [[ "$BUILD_MODE" == "local" ]]; then
    COMPOSE_FILE="docker-compose.sqlite.local.yml"
  else
    COMPOSE_FILE="docker-compose.sqlite.yml"
  fi

  if [[ ! -f "$COMPOSE_FILE" ]]; then
    error "找不到 $COMPOSE_FILE，请确认在项目根目录下运行"
  fi

  success "Compose 文件: $COMPOSE_FILE"

  COMPOSE_FILE_ARGS=(-f "$COMPOSE_FILE")
}

compose_cmd_display() {
  local display="$COMPOSE_CMD"
  local arg
  for arg in "${COMPOSE_FILE_ARGS[@]}"; do
    display+=" $arg"
  done
  printf "%s" "$display"
}

# ---------- 部署 ----------
deploy() {
  echo ""
  info "开始部署..."

  if [[ "$BUILD_MODE" == "local" ]]; then
    info "本地构建并启动..."
    $COMPOSE_CMD "${COMPOSE_FILE_ARGS[@]}" up -d --build
  else
    info "拉取最新镜像..."
    $COMPOSE_CMD "${COMPOSE_FILE_ARGS[@]}" pull
    info "启动服务..."
    $COMPOSE_CMD "${COMPOSE_FILE_ARGS[@]}" up -d
  fi

  echo ""
  success "部署完成!"
  echo ""

  local PUBLIC_IP="" LAN_IP=""

  if [[ "$BIND_MODE" == "all" ]]; then
    # 仅在对外开放时才探测/展示对外 IP
    PUBLIC_IP=$(curl -fsS4 --max-time 3 https://ifconfig.me 2>/dev/null \
      || curl -fsS4 --max-time 3 https://api.ipify.org 2>/dev/null \
      || curl -fsS4 --max-time 3 https://ipinfo.io/ip 2>/dev/null \
      || true)
    PUBLIC_IP=$(echo "$PUBLIC_IP" | tr -d '[:space:]')

    if command -v hostname >/dev/null 2>&1; then
      LAN_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || true)
    fi
    if [[ -z "$LAN_IP" ]] && command -v ip >/dev/null 2>&1; then
      LAN_IP=$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src"){print $(i+1); exit}}')
    fi
    if [[ -z "$LAN_IP" ]] && command -v ifconfig >/dev/null 2>&1; then
      LAN_IP=$(ifconfig 2>/dev/null | awk '/inet /{print $2}' | grep -v '^127\.' | head -n1)
    fi
  fi

  echo "  ┌──────────────────────────────────────────┐"
  echo "  │  部署信息                                  │"
  echo "  └──────────────────────────────────────────┘"
  echo ""
  if [[ "$BIND_MODE" == "loopback" ]]; then
    echo "  监听范围 : 127.0.0.1 (仅本机访问)"
    echo ""
    echo "  本地访问 : http://127.0.0.1:${PORT}"
    echo "             http://127.0.0.1:${PORT}/admin"
  else
    echo "  监听范围 : 0.0.0.0 (全部网络)"
    echo ""
    echo "  本地访问 : http://localhost:${PORT}"
    echo "             http://localhost:${PORT}/admin"
    if [[ -n "$LAN_IP" ]]; then
      echo "  内网访问 : http://${LAN_IP}:${PORT}"
      echo "             http://${LAN_IP}:${PORT}/admin"
    fi
    if [[ -n "$PUBLIC_IP" ]]; then
      echo "  公网访问 : http://${PUBLIC_IP}:${PORT}"
      echo "             http://${PUBLIC_IP}:${PORT}/admin"
    fi
  fi
  echo ""
  echo "  管理密钥 : ${AXISRELAY_ADMIN_SECRET}"
  echo "  查看日志 : $(compose_cmd_display) logs -f"
  echo "  停止服务 : $(compose_cmd_display) down"
  echo ""
  if [[ "$BIND_MODE" == "all" && -n "$PUBLIC_IP" ]]; then
    warn "服务对外开放，请确认防火墙/安全组已放行 ${PORT} 端口"
  fi
  if [[ "$BIND_MODE" == "loopback" ]]; then
    info "如需对外暴露，可重新运行 deploy.sh 选择「全部网络」，或在 .env 中将 AXISRELAY_BIND_HOST 改为 0.0.0.0"
  fi
  echo ""
}

# ---------- 主流程 ----------
main() {
  banner
  bootstrap_repo "$@"
  preflight
  update_repo_code
  step_deployment_route
  step_port
  step_bind
  step_database
  step_secrets
  step_build_mode
  step_confirm
  generate_env
  resolve_compose_file
  deploy
}

main "$@"
