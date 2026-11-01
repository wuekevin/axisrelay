# AxisRelay 与 james-6-23/codex2api v3.1.2 差异审计

审计日期：2026-11-01

## 基线

- AxisRelay `main`：`b4128a3acbecd1369284b3a4f75d1719046f4fbc`
- james-6-23/codex2api `main`：`dfab070b32291c0a32d62c8941970a945d6ac3d3`
- 共同祖先：`738a907de265cae6b60cae41e28b07dc7c5407c9`
- 分叉规模：AxisRelay 独有 29 个提交，James 独有 363 个提交。

这不是可安全整仓合并的普通版本差：AxisRelay 已改为 MySQL、`AXISRELAY_*` 配置命名、独立 Web/Admin 前端和公开用户体系；James 后续仍包含 SQLite/PostgreSQL、单体前端及 `CODEX2API_*` 命名。后续同步必须继续采用逻辑移植。

## 本批确认并移植

| 上游问题 | AxisRelay 审计结果 | 本批适配 |
| --- | --- | --- |
| 非流式 Codex/GPT Images 发送 HTTP 102 | 确认存在 | 非流式只等待最终 JSON；流式 SSE 和 Grok 媒体保持原行为 |
| Responses 拒绝 `configuration_update` | 确认缺失 | 加入本地输入类型白名单 |
| HTTP 代理 502/503/504 不触发 IPv6 探测 | 确认缺失 | 网关/目标不可达改为非结论错误并进入 IPv6 fallback |
| Antigravity 套餐显示名与 `plan_allow` 不匹配 | 确认缺失 | 从 entitlement tier ID 生成稳定的 `free/pro/ultra/standard/api` 键并刷新路由缓存 |
| Grok CLI 旧指纹被上游 426 拒绝 | AxisRelay 仍为 `0.2.106` | 更新为 1.0.46，补齐 Rust 架构名、UUID 请求 ID、traceparent、重复检测头 |
| Grok 长请求缺少官方缓存/压缩行为 | 确认缺失 | 保留/补齐稳定 `prompt_cache_key`，OAuth 大请求使用 zstd，保留 AxisRelay 环境变量命名 |
| OpenAI 官方价格表列移动导致错位导入 | 确认使用固定列号 | 改为按表头名称读取短/长上下文输入、缓存输入和输出 |
| `gpt-6.1-sol` 落入 Astra 兜底 | 确认缺少独立规则 | 增加独立短/长上下文价格与缓存写入价 |
| Astra 272K 以上长上下文价被清零 | 确认存在 | 恢复 $20 输入、$2 缓存读取、$75 输出及可编辑长上下文覆盖 |
| 前端构建链仍锁定 Vite 8.0.16 | 低于 James 8.3.5 | 升级声明和锁文件并重新验证 Web/Admin 双构建 |

AxisRelay 已经使用 Go 1.26.10；本批将前端从实际锁定的 Vite 8.0.16 升级到 James 使用的 8.3.5，并同步依赖锁文件。

## AxisRelay 独有能力

- MySQL 运行时和迁移体系；不回引 SQLite/PostgreSQL 运行路径。
- `frontend/web` 与 `frontend/admin` 双应用边界。
- 官网、注册登录、邮箱验证、密码找回、用户中心和公开自助门户。
- AxisRelay 模块、配置、日志和部署命名空间。
- 已移植的 Gemini 真实终态、裸 429 冷却和调度排除集稳定性修复。

## 尚未纳入本批的 James 大型能力

以下项目涉及调度、数据库、长连接状态机或前端结构，不能以单文件补丁安全移植，需单独批次和 MySQL 迁移设计：

1. 统一 Codex HTTP/WebSocket/Live attestation 与会话身份。
2. `response.interrupt` WebSocket 中断协议和终态排序。
3. Codex WebSocket quota frame 实时写入、合并持久化和账号 live polling。
4. Antigravity 定时额度刷新、耗尽门禁、平衡亲和及 quota epoch/generation 防旧写。
5. Responses 多 stream 并行、严格兼容模式和原生字段完整保留。
6. Intelligence Check 队列、历史、恢复动作与管理页面。
7. 外部调度器固定账号模式。
8. Codex Multi-Agent v2 第三方明文兼容。
9. Claude Code 会话、文档、搜索、工具发现和完整工具参数收敛。
10. James #804 的整套安全边界变更需逐项与 AxisRelay 已有请求上限、OAuth 和公开用户体系交叉审计，不能按其 SQLite/PostgreSQL 实现直接复制。

其中 `response.interrupt`、外部调度器、Multi-Agent v2、严格 Responses 兼容和完整 WS quota ingestion 的标识在当前 AxisRelay 中均不存在；attestation 目前仍是 Live 专用实现，不等同于 James 的统一出站实现。

## 后续建议顺序

1. 协议正确性：`response.interrupt`、统一 attestation、能力快照修复。
2. 配额正确性：WS quota frame、Antigravity quota scheduler/gating。
3. 安全边界：按 AxisRelay MySQL + Web/Admin + Public/Auth 架构复核 James #804。
4. 新能力：严格 Responses、多 stream、Intelligence Check、外部调度器、Multi-Agent v2。

每一批都应单独 PR；普通/race 测试、MySQL 8.0 全量测试、双前端 test/typecheck/build 全绿后再合并。
