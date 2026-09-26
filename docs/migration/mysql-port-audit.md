# S0.6 database/ MySQL Port Audit

> Sprint 0 / S0.6 仅做 `database/` 全量迁移审计，不在本阶段提前实施后续 Repository/MySQL Integration Test 或运行时重构。

## 1. 审计目标与边界

- 目标数据库：MySQL 8.0 为主要目标，同时保留 PostgreSQL 兼容性要求。
- SQLite 不作为目标运行时数据库；发现的 SQLite 专属 SQL 与 SQLite 运行路径统一归入 C 类待迁移项。
- 本阶段产物是审计清单；A/B/C 的实际 SQL 改写在后续开发阶段按文档顺序处理。
- 审计范围：`database/*.go` 的全部非测试生产文件。
- 分类扫描忽略 Go 注释，避免注释中的 BIGSERIAL/RETURNING 等字样造成误判。

## 2. A / B / C 分类标准

| 分类 | 定义 | S0.6 处理 |
| --- | --- | --- |
| A | SQL/DB API 语义完全通用，不依赖目标数据库方言。 | 记录，无需方言改写。 |
| B | SQL 语义通用，但使用 PostgreSQL 风格 `$1/$2/...` placeholder；重复或乱序参数需单独核对。 | 记录 placeholder 适配点，不在 S0.6 改运行时。 |
| C | 数据库专属语法：`ON CONFLICT`、`RETURNING`、JSON/JSONB、`ANY/ARRAY`、PG 类型/DDL/锁语法，或遗留 SQLite 专属 SQL/运行路径。 | 明确标记为待方言重写项。 |

## 3. 总览

- 生产 Go 文件：**61**
- 含数据库访问/SQL 的文件：**53**
- 无直接 SQL 的生产文件：**8**
- 审计函数/查询点：**481**
- A 类：**123**
- B 类：**209**
- C 类：**149**

### C 类特征统计

| 特征 | 查询/函数数 |
| --- | ---: |
| `ON CONFLICT` | 66 |
| `SQLite legacy` | 34 |
| `JSONB cast` | 32 |
| `RETURNING` | 29 |
| `PG DDL/type` | 20 |
| `JSONB operator` | 11 |
| `SKIP LOCKED` | 1 |

### 当前测试可追踪性

| Test 状态 | 查询/函数数 |
| --- | ---: |
| 已有直接测试引用 | 273 |
| 未发现直接测试 | 123 |
| 已有同文件测试 | 85 |

## 4. 无直接 SQL 的 production 文件

这些文件属于 `database/` 范围，但本身没有直接执行 SQL；仍计入全部 production 文件审计覆盖。

- `database/api_key_scope_limits.go`
- `database/billing.go`
- `database/continuous_retry.go`
- `database/credential_crypto.go`
- `database/credentials_accessors.go`
- `database/image_billing.go`
- `database/image_user_billing.go`
- `database/mysql_driver.go`

## 5. 全量 Query / Function 审计

| File | Query | Function | Category | Migration Status | Test |
| --- | --- | --- | :---: | --- | --- |
| `database/account_daily_usage.go` | DDL, PG DDL/type | `ensureAccountDailyUsageTable` (L121) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/account_daily_usage.go` | INSERT, UPDATE, ON CONFLICT | `UpsertAccountDailyUsage` (L191) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/account_daily_usage.go` | INSERT, UPDATE, ON CONFLICT | `UpsertAccountDailyBreakdown` (L229) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/account_daily_usage.go` | SELECT; placeholders=$1; 顺序编号 | `GetAccountDailyUsageCoverage` (L289) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/account_daily_usage.go` | DB API; placeholders=$1 / $2; 顺序编号 | `ListAccountDailyUsage` (L316) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/account_daily_usage.go` | SELECT; placeholders=$1; 顺序编号 | `SumAccountDailyUsage` (L350) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/account_daily_usage.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `SumAccountDailyUsageSince` (L390) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/account_daily_usage.go` | DELETE; placeholders=$1; 顺序编号 | `PruneAccountDailyUsage` (L414) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/account_error_fence.go` | UPDATE; placeholders=$1 / $2 / $3; 重复/乱序需核对 | `SetOwnedAccountError` (L9) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/account_error_fence.go` | UPDATE; placeholders=$1 / $2; 重复/乱序需核对 | `ClearOwnedAccountError` (L38) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/account_groups.go` | SELECT; 通用 SQL/DB API | `ListAccountGroups` (L51) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/account_groups.go` | INSERT, RETURNING | `CreateAccountGroup` (L94) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/account_groups.go` | UPDATE; 通用 SQL/DB API | `UpdateAccountGroup` (L134) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/account_groups.go` | SELECT, DELETE; placeholders=$1; 顺序编号 | `DeleteAccountGroup` (L204) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/account_groups.go` | SELECT, UPDATE, JSONB cast | `pruneDeletedGroupFromAPIKeyScopes` (L251) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/account_groups.go` | SELECT, UPDATE, JSONB cast | `pruneDeletedScopeFromAPIKeyLimits` (L300) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/account_groups.go` | INSERT, DELETE; placeholders=$1 / $2; 重复/乱序需核对 | `SetAccountGroups` (L367) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/account_groups.go` | SELECT; placeholders=$1; 顺序编号 | `GetAccountGroupIDs` (L418) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/account_groups.go` | SELECT; 通用 SQL/DB API | `ListAccountIDsInGroups` (L439) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/account_groups.go` | SELECT; 通用 SQL/DB API | `ListAccountGroupMemberships` (L470) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/account_groups.go` | SELECT; 通用 SQL/DB API | `ListAccountGroupMembershipsByAccountIDs` (L491) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/account_groups.go` | SELECT; 通用 SQL/DB API | `VerifyAccountGroupIDs` (L527) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/account_groups.go` | UPDATE, JSONB cast | `UpdateAccountTags` (L564) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/account_groups.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `UpdateAccountProxyURL` (L584) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/account_health.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `GetAccountsHealthBucketsByIDs` (L29) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/account_health.go` | SELECT, PG DDL/type | `getPostgresAccountHealthBuckets` (L112) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/account_list_projection.go` | SELECT, JSONB cast, JSONB operator, SQLite legacy | `ListAccountListProjection` (L14) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/account_list_projection.go` | SELECT; 通用 SQL/DB API | `ListActiveByIDs` (L226) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/account_page_stats.go` | SELECT; placeholders=$1; 顺序编号 | `getAccountRequestCountsByIDs` (L45) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/account_page_stats.go` | SELECT; placeholders=$1; 顺序编号 | `attachErrorStatusCounts` (L97) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/account_page_stats.go` | SELECT; placeholders=$1; 顺序编号 | `attachSuccessModelCounts` (L138) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/account_page_stats.go` | SELECT; placeholders=$1 / $2; 重复/乱序需核对 | `GetAccountUsageWindowsByIDs` (L184) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/account_page_stats.go` | SELECT; placeholders=$1; 顺序编号 | `GetAccountUsageSinceByIDs` (L230) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/account_page_stats.go` | SELECT; placeholders=$1; 顺序编号 | `GetAccountModelCountsSinceByIDs` (L260) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/antigravity_oauth_settings.go` | SELECT; 通用 SQL/DB API | `LoadAntigravityOAuthConfig` (L15) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/antigravity_oauth_settings.go` | INSERT, UPDATE, ON CONFLICT | `SaveAntigravityOAuthConfig` (L35) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/antigravity_settings.go` | SELECT; 通用 SQL/DB API | `LoadAntigravityConfig` (L14) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/antigravity_settings.go` | INSERT, UPDATE, ON CONFLICT | `SaveAntigravityConfig` (L34) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/api_key_auth_cache.go` | SELECT; 通用 SQL/DB API | `GetAPIKeyAuthRevision` (L25) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/api_key_auth_cache.go` | SELECT; placeholders=$1; 顺序编号 | `GetAPIKeyAuthQuota` (L36) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/api_key_auth_cache.go` | SELECT, INSERT, UPDATE, DELETE, DDL, ON CONFLICT, PG DDL/type | `ensureAPIKeyAuthCacheSchema` (L47) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/api_key_model_request_limits.go` | SELECT, INSERT, UPDATE, ON CONFLICT | `ConsumeAPIKeyModelRequest` (L209) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/api_key_model_request_limits.go` | SELECT; placeholders=$1 / $2 / $3; 顺序编号 | `GetAPIKeyModelRequestUsage` (L284) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/api_key_model_request_limits.go` | SELECT, UPDATE; placeholders=$1; 顺序编号 | `lockAPIKeyLimits` (L305) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/api_key_scope_counters.go` | SELECT; placeholders=$1; 顺序编号 | `ListAPIKeyScopeCounters` (L36) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/api_key_scope_counters.go` | SELECT; 通用 SQL/DB API | `ListAPIKeyScopeCountersForKeys` (L79) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/api_key_scope_counters.go` | INSERT, UPDATE, ON CONFLICT | `ResetAPIKeyScopeCounter` (L122) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/api_key_scope_counters.go` | SELECT, INSERT, UPDATE, ON CONFLICT | `applyAPIKeyScopeCountersWithExec` (L164) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/api_key_usage.go` | SELECT; placeholders=$1 / $2; 重复/乱序需核对 | `getAPIKeyUsageSince` (L49) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/api_key_usage.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `GetAPIKeyAccountWindowUsage` (L101) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/api_key_usage.go` | SELECT; placeholders=$1 / $2 / $3 / $4 / $5; 重复/乱序需核对 | `GetAPIKeyAccountWindowsUsage` (L144) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/api_key_usage.go` | SELECT; 通用 SQL/DB API | `GetAPIKeysAccountWindowUsage` (L203) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/api_key_usage.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `ListAPIKeyTokenStats` (L279) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/api_key_usage.go` | SELECT; placeholders=$1 / $2 / $3; 顺序编号 | `ListAPIKeyAccountStats` (L376) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/api_key_usage.go` | SELECT; 通用 SQL/DB API | `attachAPIKeyAccountGroups` (L464) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/api_key_usage.go` | SELECT; 通用 SQL/DB API | `ListAPIKeyLastUsedAt` (L549) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/api_key_usage.go` | SELECT; placeholders=$1; 重复/乱序需核对 | `getAllAPIKeysCostSince` (L599) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/api_key_usage.go` | SELECT; 通用 SQL/DB API | `getAPIKeySelfUsageSummary` (L874) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/api_key_usage.go` | SELECT; 通用 SQL/DB API | `listAPIKeySelfUsageBreakdown` (L911) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/api_key_usage.go` | SELECT; 通用 SQL/DB API | `listAPIKeySelfRecentLogs` (L969) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/channel_test_settings.go` | SELECT; 通用 SQL/DB API | `LoadChannelTestConfig` (L60) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/channel_test_settings.go` | INSERT, UPDATE, ON CONFLICT | `SaveChannelTestConfig` (L84) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/claude_cli_version.go` | SELECT; 通用 SQL/DB API | `GetClaudeSyncedCLIVersion` (L13) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/claude_cli_version.go` | INSERT, UPDATE, ON CONFLICT | `UpdateClaudeSyncedCLIVersion` (L29) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/claude_cli_version.go` | SELECT, UPDATE, TX, JSONB cast | `UpdateAccountCustomHeaders` (L44) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/codex_invite_recipients.go` | DB API; 通用 SQL/DB API | `Error` (L65) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/codex_invite_recipients.go` | DDL, PG DDL/type | `codexInviteRecipientDDL` (L141) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/codex_invite_recipients.go` | DB API; 通用 SQL/DB API | `ensureCodexInviteRecipientTable` (L169) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/codex_invite_recipients.go` | SELECT; placeholders=$1; 顺序编号 | `listCodexInviteRecipientsByReservation` (L216) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/codex_invite_recipients.go` | SELECT; 通用 SQL/DB API | `listCodexInviteRecipientsByKeys` (L234) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/codex_invite_recipients.go` | INSERT, ON CONFLICT | `ReserveCodexInviteRecipients` (L265) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/codex_invite_recipients.go` | UPDATE; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 重复/乱序需核对 | `FinalizeCodexInviteRecipients` (L318) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_invite_recipients.go` | UPDATE; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 顺序编号 | `MarkCodexInviteRecipientsKnownInvited` (L374) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_invite_recipients.go` | UPDATE; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 顺序编号 | `MarkCodexInviteRecipientsUnknown` (L429) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_invite_recipients.go` | DELETE; placeholders=$1 / $2; 顺序编号 | `ReleaseCodexInviteRecipients` (L456) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_invite_recipients.go` | INSERT, UPDATE, ON CONFLICT | `UpsertCodexInviteRecipientsFromTracking` (L481) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/codex_invite_snapshots.go` | DDL, JSONB cast, PG DDL/type | `ensureCodexInviteSnapshotTable` (L60) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/codex_invite_snapshots.go` | SELECT; placeholders=$1 / $2 / $3; 顺序编号 | `GetCodexInviteSnapshot` (L110) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_invite_snapshots.go` | INSERT, UPDATE, ON CONFLICT, JSONB cast | `UpsertCodexInviteSnapshot` (L145) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/codex_invite_snapshots.go` | DELETE; placeholders=$1; 顺序编号 | `DeleteCodexInviteSnapshots` (L187) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_invite_snapshots.go` | DELETE; placeholders=$1; 顺序编号 | `PurgeExpiredCodexInviteSnapshots` (L197) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_refresh.go` | DDL; 通用 SQL/DB API | `ensureCodexRefreshSchema` (L50) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/codex_refresh.go` | SELECT, UPDATE, JSONB operator, SQLite legacy | `codexRefreshRows` (L59) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有同文件测试 |
| `database/codex_refresh.go` | INSERT, TX, ON CONFLICT | `BeginCodexRefresh` (L89) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/codex_refresh.go` | SELECT, UPDATE, DELETE, TX; placeholders=$1 / $2; 重复/乱序需核对 | `FinishCodexRefresh` (L139) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_refresh.go` | DELETE, TX; placeholders=$1 / $2; 顺序编号 | `FailCodexRefresh` (L207) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/codex_refresh.go` | UPDATE, JSONB cast | `writeCodexRefreshCredentials` (L236) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/codex_turn_state_history.go` | DDL, PG DDL/type, SQLite legacy | `ensureCodexTurnStateHistorySchema` (L56) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有同文件测试 |
| `database/codex_turn_state_history.go` | INSERT, RETURNING | `StartCodexTurnStateHistory` (L92) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/codex_turn_state_history.go` | UPDATE; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 顺序编号 | `FinishCodexTurnStateHistory` (L101) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_turn_state_history.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `RecoverCodexTurnStateHistory` (L109) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_turn_state_history.go` | SELECT; 通用 SQL/DB API | `ListCodexTurnStateHistory` (L142) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/codex_turn_state_templates.go` | DDL; 通用 SQL/DB API | `ensureCodexTurnStateTemplateSchema` (L20) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/codex_turn_state_templates.go` | INSERT, UPDATE, ON CONFLICT | `SaveCodexTurnStateTemplate` (L48) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/codex_turn_state_templates.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `GetCodexTurnStateTemplate` (L57) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/codex_turn_state_templates.go` | SELECT; placeholders=$1; 顺序编号 | `ListCodexTurnStateTemplates` (L66) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/codex_turn_state_templates.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `ListCodexTurnStateRenewalCandidates` (L70) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_turn_state_templates.go` | DB API; 通用 SQL/DB API | `listCodexTurnStateTemplates` (L78) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/codex_turn_state_templates.go` | SELECT, INSERT, UPDATE, ON CONFLICT, RETURNING | `ClaimCodexTurnStateRenewal` (L97) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/codex_turn_state_templates.go` | UPDATE; placeholders=$1 / $2 / $3 / $4 / $5; 顺序编号 | `FinishCodexTurnStateRenewal` (L115) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_turn_state_templates.go` | SELECT, DELETE; 通用 SQL/DB API | `PruneCodexTurnStateRenewals` (L121) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/codex_turn_state_templates.go` | DELETE; placeholders=$1 / $2; 重复/乱序需核对 | `DeleteCodexTurnStateTemplate` (L133) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_turn_state_templates.go` | SELECT, DELETE; placeholders=$1; 重复/乱序需核对 | `PruneCodexTurnStateTemplates` (L138) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/codex_turn_state_templates.go` | SELECT; placeholders=$1 / $2 / $3; 顺序编号 | `CodexTurnStateRenewalProxyHashes` (L153) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/codex_turn_state_templates.go` | INSERT, ON CONFLICT | `RecordCodexTurnStateRenewalProxy` (L171) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/data_migrations.go` | SELECT, UPDATE, JSONB operator, SQLite legacy | `classifyAccountGroupChannels` (L68) | **C** | 待移除 SQLite 语法并改为目标方言 | 未发现直接测试 |
| `database/data_migrations.go` | SELECT, UPDATE; 通用 SQL/DB API | `backfillUsageLogChannel` (L94) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/data_migrations.go` | SELECT, UPDATE, JSONB operator, SQLite legacy | `backfillClaudeProviderData` (L126) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/data_migrations.go` | DDL, PG DDL/type | `ensureDataMigrationsTable` (L251) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/data_migrations.go` | INSERT, TX, ON CONFLICT | `runDataMigrationOnce` (L264) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/data_migrations.go` | UPDATE, JSONB cast | `migrateWorkspaceIdentityV3` (L376) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/data_migrations.go` | SELECT; 通用 SQL/DB API | `listOAuthIdentityDedupeAccounts` (L465) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/data_migrations.go` | UPDATE; 通用 SQL/DB API | `softDeleteAccountsTx` (L588) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/data_migrations.go` | INSERT; placeholders=$1 / $2; 顺序编号 | `insertAccountEventsTx` (L622) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/grok_state.go` | DDL, JSONB cast, PG DDL/type | `ensureGrokStateTables` (L144) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/grok_state.go` | SELECT, INSERT, ON CONFLICT | `ensureGrokStateHistoricalBackfill` (L278) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/grok_state.go` | SELECT, INSERT, UPDATE, TX, ON CONFLICT | `runGrokStateHistoricalBackfillBatch` (L305) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/grok_state.go` | SELECT, UPDATE; placeholders=$1 / $2; 顺序编号 | `backfillCredentialFamilies` (L423) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/grok_state.go` | SELECT, UPDATE; placeholders=$1 / $2; 重复/乱序需核对 | `backfillCredentialFamiliesBatch` (L457) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/grok_state.go` | SELECT, INSERT, UPDATE, TX, ON CONFLICT, RETURNING, JSONB cast | `InsertGrokAccountIfAbsent` (L537) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/grok_state.go` | SELECT, INSERT, UPDATE, TX, ON CONFLICT, JSONB cast | `ReauthGrokAccount` (L616) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/grok_state.go` | SELECT, INSERT, ON CONFLICT | `backfillGrokCredentialIdentityClaims` (L709) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/grok_state.go` | SELECT, INSERT, ON CONFLICT, JSONB operator | `backfillGrokCredentialIdentityClaimsBatch` (L746) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/grok_state.go` | SELECT; placeholders=$1; 顺序编号 | `GetAccountCredentialState` (L794) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/grok_state.go` | SELECT, UPDATE; placeholders=$1 / $2; 重复/乱序需核对 | `EnsureAccountCredentialFamilyID` (L801) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/grok_state.go` | SELECT, UPDATE, TX, JSONB cast | `UpdateAccountCredentialsCAS` (L819) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/grok_state.go` | SELECT, UPDATE, TX, JSONB cast | `ReplaceAccountCredentialsCAS` (L898) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/grok_state.go` | SELECT, UPDATE, TX, JSONB cast | `MergeAccountCredentialsForGeneration` (L966) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/grok_state.go` | SELECT, UPDATE; placeholders=$1; 顺序编号 | `verifyGrokGenerationTx` (L1035) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/grok_state.go` | INSERT, UPDATE, TX, ON CONFLICT, JSONB cast | `UpsertGrokAccountFact` (L1054) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/grok_state.go` | INSERT, UPDATE, TX, ON CONFLICT, JSONB cast | `UpsertGrokAccountFactAndExpireCapabilities` (L1108) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/grok_state.go` | SELECT, INSERT, UPDATE, DELETE, TX, ON CONFLICT, JSONB cast | `ReplaceGrokModelCatalog` (L1165) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/grok_state.go` | SELECT, UPDATE; placeholders=$1 / $2 / $3 / $4 / $5; 重复/乱序需核对 | `TouchGrokModelCatalogNotModified` (L1310) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/grok_state.go` | UPDATE, TX; placeholders=$1 / $2 / $3 / $4 / $5; 重复/乱序需核对 | `UpdateGrokModelsETagHint` (L1328) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/grok_state.go` | INSERT, UPDATE, TX, ON CONFLICT | `UpsertGrokModelCapability` (L1382) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/grok_state.go` | UPDATE, TX; placeholders=$1 / $2 / $3; 顺序编号 | `ExpireGrokModelCapabilities` (L1425) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/grok_state.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `GetGrokAccountFact` (L1572) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/grok_state.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `GetGrokModelCatalog` (L1594) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/grok_state.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `getGrokCatalogItems` (L1612) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/grok_state.go` | SELECT; placeholders=$1; 顺序编号 | `GetGrokModelCapabilities` (L1620) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/grok_state.go` | SELECT; placeholders=$1; 重复/乱序需核对 | `GetGrokAccountState` (L1641) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/grok_state.go` | DELETE; 通用 SQL/DB API | `deleteGrokAccountStateTx` (L1721) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/helpers.go` | DB API; 通用 SQL/DB API | `insertRowID` (L417) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/image_job_result.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `GetImageGenerationJobResult` (L7) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/image_studio.go` | INSERT, RETURNING | `InsertImagePromptTemplate` (L169) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/image_studio.go` | UPDATE; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 顺序编号 | `UpdateImagePromptTemplate` (L179) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | DELETE; placeholders=$1; 顺序编号 | `DeleteImagePromptTemplate` (L189) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | SELECT; placeholders=$1; 顺序编号 | `GetImagePromptTemplate` (L194) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | SELECT; 通用 SQL/DB API | `ListImagePromptTemplates` (L210) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/image_studio.go` | UPDATE; placeholders=$1; 顺序编号 | `IncrementImagePromptTemplateUsage` (L284) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | INSERT, RETURNING | `InsertImageGenerationJob` (L296) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/image_studio.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `MarkImageJobRunning` (L309) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | UPDATE; placeholders=$1 / $2 / $3; 顺序编号 | `MarkImageJobSucceeded` (L318) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | UPDATE; placeholders=$1 / $2 / $3 / $4; 顺序编号 | `MarkImageJobSucceededWithWarning` (L327) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/image_studio.go` | UPDATE; placeholders=$1 / $2 / $3 / $4; 顺序编号 | `MarkImageJobFailed` (L339) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `UpdateImageGenerationJobParamsJSON` (L351) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | UPDATE; placeholders=$1 / $2 / $3 / $4; 顺序编号 | `MarkInterruptedImageJobs` (L360) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | SELECT; placeholders=$1; 顺序编号 | `GetImageGenerationJob` (L369) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | SELECT; placeholders=$1 / $2 / $3; 重复/乱序需核对 | `ListImageGenerationJobs` (L389) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | DELETE, TX; placeholders=$1; 重复/乱序需核对 | `DeleteImageGenerationJob` (L447) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | INSERT, RETURNING | `InsertImageAsset` (L510) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/image_studio.go` | DB API; placeholders=$1; 顺序编号 | `GetImageAsset` (L525) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | SELECT; placeholders=$1 / $2 / $3; 重复/乱序需核对 | `ListImageAssets` (L540) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | DB API; placeholders=$1; 顺序编号 | `ListImageAssetsByJobID` (L579) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/image_studio.go` | SELECT; placeholders=$1; 顺序编号 | `GetImageAssetJobAPIKeyID` (L593) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | DELETE; placeholders=$1; 顺序编号 | `DeleteImageAsset` (L607) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | SELECT; 通用 SQL/DB API | `imageAssetSelectSQL` (L614) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/image_studio.go` | SELECT; placeholders=$1; 顺序编号 | `GetAPIKeyByID` (L672) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/image_studio.go` | SELECT; 通用 SQL/DB API | `FirstAPIKey` (L684) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/invite_guide_settings.go` | SELECT; 通用 SQL/DB API | `LoadInviteGuideConfig` (L36) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/invite_guide_settings.go` | INSERT, UPDATE, ON CONFLICT | `SaveInviteGuideConfig` (L60) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/maintenance_jobs.go` | SELECT, INSERT, ON CONFLICT, JSONB operator, SQLite legacy | `SeedGrokMaintenanceJobs` (L27) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有同文件测试 |
| `database/maintenance_jobs.go` | SELECT, UPDATE, TX, RETURNING, SKIP LOCKED | `ClaimMaintenanceJobs` (L63) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/maintenance_jobs.go` | UPDATE; placeholders=$1 / $2 / $3 / $4; 顺序编号 | `CompleteMaintenanceJob` (L150) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/maintenance_jobs.go` | UPDATE; placeholders=$1 / $2 / $3 / $4 / $5; 顺序编号 | `FailMaintenanceJob` (L169) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/maintenance_jobs.go` | DELETE; placeholders=$1 / $2; 顺序编号 | `DeleteMaintenanceJob` (L186) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/model_capabilities.go` | SELECT, INSERT, UPDATE, ON CONFLICT | `SaveModelCapabilities` (L28) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/model_capabilities.go` | SELECT; 通用 SQL/DB API | `ListModelCapabilities` (L91) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/model_cooldown_settings.go` | SELECT; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 顺序编号 | `GetModelCooldownSettings` (L111) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/model_cooldown_settings.go` | INSERT, UPDATE, ON CONFLICT | `UpdateModelCooldownSettings` (L143) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/model_pricing_override.go` | SELECT, DELETE; 通用 SQL/DB API | `MutateModelPricingSettings` (L191) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/models.go` | SELECT; 通用 SQL/DB API | `ListModelRegistry` (L29) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/models.go` | INSERT, UPDATE, ON CONFLICT | `UpsertModelRegistryRows` (L75) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/models.go` | DELETE; placeholders=$1; 顺序编号 | `DeleteModelRegistryRows` (L132) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/models.go` | SELECT; 通用 SQL/DB API | `GetModelRegistrySyncState` (L145) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/models.go` | INSERT, UPDATE, ON CONFLICT | `UpdateModelRegistrySyncState` (L168) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/official_pricing_sync.go` | INSERT, DDL, ON CONFLICT | `ensureOfficialPricingSyncConfig` (L42) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/official_pricing_sync.go` | SELECT; 通用 SQL/DB API | `GetOfficialPricingSyncConfig` (L76) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/official_pricing_sync.go` | UPDATE; placeholders=$1 / $2 / $3 / $4 / $5; 顺序编号 | `UpdateOfficialPricingSyncConfig` (L94) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/official_pricing_sync.go` | UPDATE; placeholders=$1 / $2 / $3 / $4; 重复/乱序需核对 | `RecordOfficialPricingSyncResult` (L111) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | DB API; 通用 SQL/DB API | `GetCredentialStringSlice` (L182) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, DDL, ON CONFLICT, SQLite legacy | `New` (L372) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `ensureUsageLogsGenerationIndex` (L539) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT, DDL, PG DDL/type | `ensureUsageLogsOnlineIndex` (L574) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/postgres.go` | DDL; 通用 SQL/DB API | `ensureUsageStatsBaselineBillingColumns` (L596) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT, DDL; 通用 SQL/DB API | `ensureUsageStatsRollup` (L660) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT, INSERT, UPDATE, DELETE, TX, ON CONFLICT | `rebuildUsageStatsRollup` (L706) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; placeholders=$1; 顺序编号 | `loadUsageStatsRollup` (L753) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | SELECT, INSERT, UPDATE, ON CONFLICT | `applyUsageStatsRollupWithExec` (L784) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `drainBackgroundTasks` (L906) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `notifyLogFlush` (L1041) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT, UPDATE, DELETE, DDL, JSONB cast, JSONB operator, PG DDL/type | `migrate` (L1052) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `ListAPIKeys` (L1922) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `CountAPIKeys` (L1941) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; placeholders=$1; 顺序编号 | `GetAPIKeyByValue` (L1950) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, RETURNING, JSONB cast | `InsertAPIKeyWithOptions` (L1967) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `UpdateAPIKeyName` (L2006) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `UpdateAPIKeyQuotaLimit` (L2022) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `UpdateAPIKeyExpiresAt` (L2041) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | UPDATE, JSONB cast | `UpdateAPIKeyAllowedGroups` (L2058) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE, JSONB cast | `UpdateAPIKeyLimits` (L2088) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE, JSONB cast | `UpdateAPIKey` (L2109) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE, RETURNING | `ResetAPIKeyQuota` (L2209) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE, RETURNING | `ResetAllAPIKeyQuotas` (L2228) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `GetSystemSettings` (L2555) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, UPDATE, TX, ON CONFLICT | `UpdateContinuousRetryPolicy` (L2805) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT, UPDATE; 通用 SQL/DB API | `continuousRetryPolicySelectQuery` (L2877) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, UPDATE, ON CONFLICT | `UpdateSystemSettings` (L2908) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, UPDATE, ON CONFLICT | `UpdateCodexSyncedCLIVersion` (L3199) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, UPDATE, ON CONFLICT | `UpdateModelsListReadMaxBytes` (L3210) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, UPDATE, ON CONFLICT | `UpdateModelPricingSettings` (L3224) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, UPDATE, ON CONFLICT | `UpdateClaudeConfig` (L3417) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/postgres.go` | DELETE, TX; placeholders=$1; 重复/乱序需核对 | `DeleteAPIKey` (L3440) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `GetAllAPIKeyValues` (L3473) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | UPDATE, TX; placeholders=$1 / $2; 顺序编号 | `SetAccountProxyURLs` (L3520) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `CountAccountsByProxyURL` (L3549) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `ListProxies` (L3576) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; placeholders=$1; 顺序编号 | `GetProxy` (L3612) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `ListProxiesByIDs` (L3643) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `ListEnabledProxies` (L3688) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, ON CONFLICT, RETURNING | `InsertProxy` (L3716) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, ON CONFLICT, RETURNING | `InsertProxies` (L3725) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/postgres.go` | DELETE; placeholders=$1; 顺序编号 | `DeleteProxy` (L3757) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | DELETE; 通用 SQL/DB API | `DeleteProxies` (L3773) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT, UPDATE; placeholders=$1; 顺序编号 | `UpdateProxy` (L3794) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 顺序编号 | `UpdateProxyTestResult` (L3844) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT, UPDATE, DELETE, TX, RETURNING | `CleanErrorProxies` (L3880) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE, RETURNING | `unbindAccountsFromProxyURLsTx` (L4003) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT, UPDATE, DELETE, TX, RETURNING | `RetireProxiesByIDs` (L4027) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE, RETURNING | `RebindAccountProxyURLs` (L4116) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `startLogFlusher` (L4535) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `stopTimer` (L4555) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | INSERT; 通用 SQL/DB API | `salvageUsageLogBatchWith` (L4702) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, TX; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 顺序编号 | `insertSQLiteUsageLogBatch` (L4779) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | TX; 通用 SQL/DB API | `batchInsertLogs` (L4837) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | INSERT; 通用 SQL/DB API | `batchInsertLogsChunk` (L4882) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | UPDATE; placeholders=$1 / $2; 重复/乱序需核对 | `applyAPIKeyQuotaUsageWithExec` (L4954) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | SELECT; placeholders=$1 / $2 / $3; 重复/乱序需核对 | `getUsageStats` (L5084) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | SELECT; placeholders=$1; 顺序编号 | `CountTodayRequestsByChannel` (L5201) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `getUsageModelStats` (L5252) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `populateUsageBreakdownStats` (L5309) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `getUsageEndpointStats` (L5356) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `getUsageAPIKeyStats` (L5399) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `GetTrafficSnapshot` (L5444) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; placeholders=$1; 顺序编号 | `ListRecentUsageLogs` (L5490) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; placeholders=$1 / $2 / $3 / $4; 重复/乱序需核对 | `GetChartAggregation` (L5653) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; placeholders=$1 / $2 / $3; 重复/乱序需核对 | `GetAccountUsageStats` (L5746) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `ListUsageLogsByTimeRange` (L5972) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `usageLogDimensionWhere` (L6130) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `GetUsageErrorSummary` (L6297) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `ListUsageLogsByTimeRangePaged` (L6337) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `ListUsageLogsByFilter` (L6410) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT, UPDATE, DELETE, TX, PG DDL/type, SQLite legacy | `ClearUsageLogs` (L6472) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `currentAccountUsageGenerationPredicate` (L6587) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; placeholders=$1; 顺序编号 | `GetAccountRequestCounts` (L6593) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; placeholders=$1; 顺序编号 | `GetAccountTimeRangeUsage` (L6638) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; placeholders=$1 / $2; 重复/乱序需核对 | `GetAccountUsageWindows` (L6668) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `GetAccountBilledSince` (L6706) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT, PG DDL/type | `getAccountsBilledSinceChunk` (L6748) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT, JSONB operator, SQLite legacy | `ListActiveByChannel` (L6806) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; placeholders=$1; 顺序编号 | `ListActiveModelCooldowns` (L6880) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `ListActiveModelCooldownsForAccounts` (L6926) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, UPDATE, ON CONFLICT | `SetModelCooldown` (L6977) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | DELETE; placeholders=$1 / $2; 顺序编号 | `ClearModelCooldown` (L7008) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | DELETE; placeholders=$1; 顺序编号 | `ClearAllModelCooldowns` (L7017) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | DELETE; placeholders=$1; 顺序编号 | `ClearExpiredModelCooldowns` (L7025) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | SELECT; placeholders=$1; 顺序编号 | `getAccountByID` (L7041) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | SELECT, UPDATE, TX, JSONB cast | `UpdateAccountSchedulerConfig` (L7107) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT, INSERT, UPDATE, DELETE, TX, JSONB cast | `UpdateAccountSchedulerMetadata` (L7197) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE, TX; 通用 SQL/DB API | `BatchUpdateAccountMetadata` (L7307) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT, UPDATE; 通用 SQL/DB API | `selectBatchAccounts` (L7363) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | UPDATE, JSONB cast | `batchUpdateAccountColumns` (L7410) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/postgres.go` | UPDATE, JSONB cast | `batchUpdateAccountCredentials` (L7465) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/postgres.go` | INSERT, DELETE; placeholders=$1 / $2; 顺序编号 | `batchReplaceAccountGroups` (L7493) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `SetAccountEnabled` (L7546) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `SetAccountLocked` (L7562) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `UpdateAccountNote` (L7568) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | UPDATE, JSONB cast | `SetAccountTags` (L7584) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/postgres.go` | UPDATE; 通用 SQL/DB API | `UpdateAccountCredit` (L7598) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT, UPDATE, TX, JSONB cast | `updateCredentialsReadMerge` (L7645) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/postgres.go` | UPDATE; 通用 SQL/DB API | `updateCredentialsSQLite` (L7681) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT, UPDATE, TX; placeholders=$1 / $2; 重复/乱序需核对 | `updateCredentialsReadMergeSQLiteUnlocked` (L7738) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | SELECT, UPDATE, DELETE, TX, JSONB cast | `UpdateOpenAIResponsesAccount` (L7821) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT, UPDATE, TX, JSONB cast | `UpdateOAuthAccountCredentials` (L7873) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/postgres.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `SetError` (L7936) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE; placeholders=$1; 顺序编号 | `BatchSetError` (L7945) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | UPDATE, TX; placeholders=$1; 顺序编号 | `SoftDeleteAccount` (L7975) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `ListDeleted` (L8013) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE; placeholders=$1; 顺序编号 | `RestoreAccount` (L8087) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | DELETE, TX; placeholders=$1; 重复/乱序需核对 | `PurgeAccount` (L8113) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | SELECT, DELETE, TX; 通用 SQL/DB API | `PurgeDeletedAccounts` (L8146) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE; 通用 SQL/DB API | `BatchSoftDeleteAccounts` (L8180) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT; placeholders=$1 / $2 / $3 / $4; 重复/乱序需核对 | `BatchInsertAccountEvents` (L8217) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | UPDATE; placeholders=$1; 顺序编号 | `ClearError` (L8266) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | UPDATE; placeholders=$1 / $2 / $3; 顺序编号 | `SetCooldown` (L8275) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE; placeholders=$1 / $2 / $3 / $4; 顺序编号 | `SetCooldownWithError` (L8284) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/postgres.go` | UPDATE; placeholders=$1; 顺序编号 | `ClearCooldown` (L8293) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | INSERT, RETURNING | `InsertAccount` (L8302) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `CountAll` (L8319) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `GetAllRefreshTokens` (L8326) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | INSERT, RETURNING | `InsertATAccount` (L8348) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/postgres.go` | INSERT, RETURNING | `InsertAccountWithCredentials` (L8365) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, RETURNING | `InsertOpenAIResponsesAccount` (L8382) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | INSERT, RETURNING | `InsertAccountWithUpstream` (L8401) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT, UPDATE, TX, SQLite legacy | `insertAccountRowWithFamily` (L8425) | **C** | 待移除 SQLite 语法并改为目标方言 | 未发现直接测试 |
| `database/postgres.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `UpdateAccountName` (L8460) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `GetAllAccessTokens` (L8477) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `GetAllChatGPTAccountIDs` (L8500) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT, JSONB operator, SQLite legacy | `FindActiveAccountByOAuthIdentity` (L8526) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/postgres.go` | SELECT, JSONB operator, SQLite legacy | `FindActiveAccountByOAuthRouteIdentity` (L8583) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `GetAllOpenAIAPIKeys` (L8640) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `GetAllSessionTokens` (L8663) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/postgres.go` | INSERT; placeholders=$1 / $2 / $3; 顺序编号 | `InsertAccountEvent` (L8687) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/postgres.go` | SELECT; 通用 SQL/DB API | `InsertAccountEventAsync` (L8696) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/postgres.go` | SELECT; placeholders=$1 / $2 / $3; 顺序编号 | `GetAccountEventTrend` (L8719) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/prompt_conversation_lock.go` | SELECT, DDL, PG DDL/type, SQLite legacy | `ensurePromptConversationLocksTable` (L96) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/prompt_conversation_lock.go` | INSERT, UPDATE, ON CONFLICT, RETURNING | `LockPromptConversation` (L222) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/prompt_conversation_lock.go` | DB API; placeholders=$1; 顺序编号 | `GetPromptConversationLock` (L264) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_conversation_lock.go` | DB API; placeholders=$1; 顺序编号 | `GetActivePromptConversationLock` (L271) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_conversation_lock.go` | DB API; placeholders=$1 / $2 / $3 / $4 / $5; 重复/乱序需核对 | `GetActivePromptConversationRestriction` (L293) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_conversation_lock.go` | DB API; placeholders=$1; 顺序编号 | `GetActivePromptConversationLockBySessionHash` (L325) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_conversation_lock.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `HasActivePromptFingerprintReplayLocks` (L336) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_conversation_lock.go` | UPDATE; placeholders=$1 / $2 / $3 / $4; 重复/乱序需核对 | `expirePromptConversationLockIfNeeded` (L364) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_conversation_lock.go` | UPDATE; placeholders=$1 / $2 / $3; 重复/乱序需核对 | `UnlockPromptConversation` (L390) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_conversation_lock.go` | UPDATE, RETURNING | `UnlockPromptConversationUserCooldown` (L420) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/prompt_filter.go` | SELECT; 通用 SQL/DB API | `close` (L103) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_filter.go` | SELECT; 通用 SQL/DB API | `enqueueJob` (L180) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/prompt_filter.go` | SELECT; 通用 SQL/DB API | `worker` (L242) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_filter.go` | SELECT; 通用 SQL/DB API | `next` (L304) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_filter.go` | SELECT; 通用 SQL/DB API | `nextDraining` (L320) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/prompt_filter.go` | SELECT; 通用 SQL/DB API | `WaitPromptFilterAuditIdle` (L560) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_filter.go` | INSERT, TX, RETURNING | `InsertPromptFilterLog` (L686) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/prompt_filter.go` | SELECT; 通用 SQL/DB API | `ListPromptFilterLogsPage` (L758) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_filter.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `FindNearestPromptFilterLog` (L890) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_filter.go` | DELETE, SQLite legacy | `ClearPromptFilterLogs` (L969) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/prompt_filter.go` | DELETE; placeholders=$1; 顺序编号 | `ClearPromptFilterLogsByReviewStatus` (L984) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_filter.go` | DELETE; placeholders=$1; 顺序编号 | `ClearPromptFilterLogsBySource` (L992) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_filter_newapi_binding.go` | UPDATE, DDL; 通用 SQL/DB API | `ensurePromptFilterNewAPIBindingsTable` (L82) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_filter_newapi_binding.go` | SELECT; 通用 SQL/DB API | `ListPromptFilterNewAPIBindings` (L151) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_filter_newapi_binding.go` | SELECT; placeholders=$1; 顺序编号 | `GetPromptFilterNewAPIBinding` (L168) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_filter_newapi_binding.go` | INSERT; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 顺序编号 | `CreatePromptFilterNewAPIBinding` (L172) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_filter_newapi_binding.go` | UPDATE; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 顺序编号 | `UpdatePromptFilterNewAPIBinding` (L205) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_filter_newapi_binding.go` | UPDATE; placeholders=$1 / $2 / $3; 重复/乱序需核对 | `ReplacePromptFilterNewAPIBindingSecretAt` (L252) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_filter_newapi_binding.go` | DELETE; placeholders=$1; 顺序编号 | `DeletePromptFilterNewAPIBinding` (L274) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_filter_review_keys.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `CompareAndSwapPromptFilterReviewAPIKeys` (L14) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_incident_subjects.go` | SELECT; placeholders=$1; 顺序编号 | `ListPromptRiskSubjectsForIncident` (L27) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_intelligence_ai.go` | SELECT; 通用 SQL/DB API | `ListLatestPromptRuleCandidateAIAnalyses` (L27) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_intelligence_ai.go` | SELECT, UPDATE, TX; placeholders=$1 / $2; 重复/乱序需核对 | `ReconcilePromptRuleCandidateIdentityStatuses` (L123) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_intelligence_ai.go` | SELECT, INSERT, UPDATE, TX, ON CONFLICT | `AddPromptRuleCandidateEvidence` (L162) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/prompt_intelligence_ai.go` | SELECT; placeholders=$1; 顺序编号 | `GetPromptRuleCandidateEvidence` (L235) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_intelligence_ai.go` | SELECT, INSERT, UPDATE, TX, ON CONFLICT | `CompareAndSwapPromptFilterAdvancedConfigWithEvidence` (L262) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/prompt_policy_incident.go` | UPDATE, DDL, PG DDL/type, SQLite legacy | `ensurePromptPolicyIncidentsTable` (L182) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/prompt_policy_incident.go` | SELECT, INSERT, ON CONFLICT | `migrateLegacyPromptPolicyIncidents` (L314) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/prompt_policy_incident.go` | SELECT; placeholders=$1; 顺序编号 | `loadPromptPolicyShadowEvidenceTx` (L430) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_policy_incident.go` | SELECT, UPDATE; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 重复/乱序需核对 | `reconcileStoredPromptPolicyIncidentFromShadowTx` (L472) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_policy_incident.go` | DELETE; 通用 SQL/DB API | `mergePromptPolicyCandidateEvidenceMetadata` (L535) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_policy_incident.go` | INSERT, UPDATE, TX, ON CONFLICT | `PersistPromptPolicyIncident` (L648) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/prompt_policy_incident.go` | DB API; placeholders=$1; 顺序编号 | `GetPromptPolicyIncident` (L739) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_policy_incident.go` | SELECT, UPDATE, DELETE, TX; placeholders=$1; 重复/乱序需核对 | `DeletePromptPolicyIncident` (L749) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_policy_incident.go` | SELECT; 通用 SQL/DB API | `ListPromptPolicyIncidentsPage` (L785) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_policy_incident.go` | UPDATE, DELETE, TX, SQLite legacy | `ClearPromptPolicyIncidents` (L907) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/prompt_retention.go` | INSERT, DDL, ON CONFLICT | `ensurePromptLogRetentionConfig` (L71) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/prompt_retention.go` | SELECT; 通用 SQL/DB API | `GetPromptLogRetentionConfig` (L100) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_retention.go` | UPDATE; placeholders=$1; 顺序编号 | `UpdatePromptLogRetentionDays` (L118) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_retention.go` | SELECT, UPDATE; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 顺序编号 | `RecordPromptLogRetentionRun` (L129) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_retention.go` | SELECT, DELETE; placeholders=$1; 重复/乱序需核对 | `purgePromptLogs` (L174) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_retention.go` | SELECT, DELETE; placeholders=$1 / $2; 顺序编号 | `purgeOrphanPromptRiskSources` (L234) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_retention.go` | SELECT; 通用 SQL/DB API | `purgeInBatches` (L245) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/prompt_retention.go` | SELECT, DELETE; placeholders=$1 / $2; 重复/乱序需核对 | `deletePromptIncidentEvidenceTx` (L291) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_retention.go` | SELECT, DELETE; 通用 SQL/DB API | `deleteAllPromptIncidentEvidenceTx` (L312) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/prompt_review_profiles.go` | DDL; 通用 SQL/DB API | `ensurePromptReviewProfiles` (L28) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/prompt_review_profiles.go` | SELECT; 通用 SQL/DB API | `ListPromptReviewProfiles` (L49) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_review_profiles.go` | SELECT; placeholders=$1; 顺序编号 | `GetPromptReviewProfile` (L74) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_review_profiles.go` | INSERT, UPDATE, ON CONFLICT | `UpsertPromptReviewProfile` (L95) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/prompt_review_profiles.go` | UPDATE; placeholders=$1; 顺序编号 | `SetPromptReviewProfileActive` (L121) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_review_profiles.go` | DELETE; placeholders=$1; 顺序编号 | `DeletePromptReviewProfile` (L145) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_risk_profile.go` | DDL, PG DDL/type, SQLite legacy | `ensurePromptRiskEventsTable` (L247) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/prompt_risk_profile.go` | INSERT, UPDATE, ON CONFLICT | `upsertPromptRiskIdentity` (L520) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/prompt_risk_profile.go` | TX; 通用 SQL/DB API | `UpsertPromptRiskIdentities` (L543) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_risk_profile.go` | DB API; 通用 SQL/DB API | `shortRiskKey` (L605) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/prompt_risk_profile.go` | INSERT, ON CONFLICT | `insertPromptRiskSignal` (L617) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/prompt_risk_profile.go` | SELECT, UPDATE; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 重复/乱序需核对 | `reconcilePromptRiskReviewForSubject` (L658) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_risk_profile.go` | SELECT, TX; placeholders=$1 / $2; 顺序编号 | `backfillPromptRiskLogs` (L721) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_risk_profile.go` | SELECT, TX; placeholders=$1 / $2; 顺序编号 | `backfillPromptRiskIncidents` (L786) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_risk_profile.go` | SELECT; placeholders=$1; 顺序编号 | `loadPromptRiskIdentities` (L827) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_risk_profile.go` | SELECT; placeholders=$1; 顺序编号 | `listPromptRiskIdentities` (L875) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_risk_profile.go` | SELECT; 通用 SQL/DB API | `promptRiskActiveRestrictionSubjects` (L948) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/prompt_risk_profile.go` | SELECT; placeholders=$1 / $2 / $3 / $4; 重复/乱序需核对 | `ListPromptRiskProfiles` (L1080) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_risk_profile.go` | SELECT; placeholders=$1 / $2 / $3 / $4; 重复/乱序需核对 | `ListPromptRiskEvents` (L1424) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_risk_profile.go` | DELETE, SQLite legacy | `ClearPromptRiskEvents` (L1481) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有同文件测试 |
| `database/prompt_risk_trust.go` | DDL, PG DDL/type, SQLite legacy | `ensurePromptRiskTrustTables` (L94) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/prompt_risk_trust.go` | SELECT, INSERT, UPDATE, TX, ON CONFLICT | `UpsertPromptRiskTrustPolicy` (L231) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/prompt_risk_trust.go` | INSERT; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 顺序编号 | `insertPromptRiskTrustEvent` (L278) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_risk_trust.go` | SELECT; 通用 SQL/DB API | `scanPromptRiskTrustPolicy` (L286) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/prompt_risk_trust.go` | DB API; placeholders=$1 / $2; 顺序编号 | `GetPromptRiskTrustPolicy` (L332) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_risk_trust.go` | SELECT; 通用 SQL/DB API | `ListPromptRiskTrustPolicies` (L339) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/prompt_risk_trust.go` | SELECT, UPDATE, TX; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 重复/乱序需核对 | `transitionPromptRiskTrustPolicy` (L399) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_risk_trust.go` | UPDATE; placeholders=$1 / $2 / $3; 顺序编号 | `ReconcilePromptRiskTrustPolicies` (L452) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_risk_trust.go` | SELECT; placeholders=$1 / $2 / $3; 顺序编号 | `GetPromptRiskAdaptiveReviewBasis` (L501) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_risk_trust.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `promptRiskTrustEvidenceSummaries` (L558) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_risk_trust.go` | SELECT; placeholders=$1 / $2 / $3; 顺序编号 | `promptRiskTrustEvidenceSummarySince` (L600) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_risk_trust.go` | UPDATE, TX; placeholders=$1; 顺序编号 | `RecordPromptRiskTrustBypass` (L696) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_risk_trust.go` | UPDATE, TX; placeholders=$1; 顺序编号 | `RecordPromptRiskTrustModelReview` (L722) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_risk_trust.go` | SELECT; placeholders=$1 / $2 / $3; 顺序编号 | `ListPromptRiskTrustEvents` (L748) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_risk_trust.go` | SELECT; placeholders=$1 / $2 / $3 / $4; 重复/乱序需核对 | `ListPromptRiskTrustEventsPage` (L778) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_rule_candidate.go` | DDL; 通用 SQL/DB API | `ensurePromptRuleCandidatesTable` (L191) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/prompt_rule_candidate.go` | TX; 通用 SQL/DB API | `StagePromptRuleCandidate` (L291) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_rule_candidate.go` | SELECT, INSERT, UPDATE, ON CONFLICT | `stagePromptRuleCandidateTx` (L326) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/prompt_rule_candidate.go` | DB API; placeholders=$1; 顺序编号 | `GetPromptRuleCandidate` (L382) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_rule_candidate.go` | DB API; placeholders=$1; 顺序编号 | `GetPromptRuleCandidateByFingerprint` (L386) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_rule_candidate.go` | SELECT; 通用 SQL/DB API | `ListPromptRuleCandidates` (L390) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/prompt_rule_candidate.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `ListPromptRuleCandidateEvidence` (L444) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_rule_candidate.go` | SELECT; placeholders=$1 / $2 / $3; 顺序编号 | `HasPromptRuleCandidateEvidence` (L478) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/prompt_rule_candidate.go` | SELECT, UPDATE, TX; placeholders=$1; 重复/乱序需核对 | `DismissPromptRuleCandidate` (L487) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/prompt_rule_candidate.go` | SELECT, INSERT, UPDATE, TX, ON CONFLICT | `PublishPromptRuleCandidate` (L529) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/prompt_rule_candidate.go` | INSERT, UPDATE, ON CONFLICT | `ReplacePromptFilterCustomPatterns` (L703) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/prompt_rule_candidate.go` | SELECT, INSERT, UPDATE, TX, ON CONFLICT | `CompareAndSwapPromptFilterCustomPatternsWithMigrationCompletions` (L728) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/proxy_risk_scoring.go` | DDL, PG DDL/type, SQLite legacy | `NormalizeProxyRiskScoringProfile` (L71) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/proxy_risk_scoring.go` | DB API; 通用 SQL/DB API | `ensureProxyRiskScoringTables` (L290) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/proxy_risk_scoring.go` | INSERT, RETURNING | `CreateProxyRiskScoringProfile` (L330) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/proxy_risk_scoring.go` | UPDATE; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 顺序编号 | `UpdateProxyRiskScoringProfile` (L358) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/proxy_risk_scoring.go` | DELETE; placeholders=$1; 顺序编号 | `DeleteProxyRiskScoringProfile` (L380) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/proxy_risk_scoring.go` | SELECT; 通用 SQL/DB API | `ListProxyRiskScoringProfiles` (L400) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/proxy_risk_scoring.go` | SELECT; placeholders=$1; 顺序编号 | `GetProxyRiskScoringProfile` (L417) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/proxy_risk_scoring.go` | SELECT, UPDATE; placeholders=$1 / $2; 重复/乱序需核对 | `ReserveProxyRiskScoringCheck` (L465) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/proxy_risk_scoring.go` | UPDATE; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 顺序编号 | `UpdateProxyRiskScoringQuota` (L494) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/proxy_risk_scoring.go` | INSERT, RETURNING | `InsertProxyRiskScoreSnapshot` (L522) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/proxy_risk_scoring.go` | SELECT; 通用 SQL/DB API | `ListLatestProxyRiskScores` (L586) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/proxy_risk_scoring.go` | SELECT; placeholders=$1 / $2 / $3 / $4; 重复/乱序需核对 | `ListProxyRiskScoreHistory` (L613) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/quality_test_prompts.go` | DDL, PG DDL/type, SQLite legacy | `ensureQualityTestPromptSchema` (L24) | **C** | 待移除 SQLite 语法并改为目标方言 | 未发现直接测试 |
| `database/quality_test_prompts.go` | SELECT; 通用 SQL/DB API | `ListQualityTestPrompts` (L47) | **A** | 无需方言改写 | 未发现直接测试 |
| `database/quality_test_prompts.go` | SELECT; placeholders=$1; 顺序编号 | `GetQualityTestPrompt` (L64) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/quality_test_prompts.go` | INSERT, RETURNING | `InsertQualityTestPrompt` (L79) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 未发现直接测试 |
| `database/quality_test_prompts.go` | UPDATE; placeholders=$1 / $2 / $3; 顺序编号 | `UpdateQualityTestPrompt` (L86) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/quality_test_prompts.go` | DELETE; placeholders=$1; 顺序编号 | `DeleteQualityTestPrompt` (L97) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/quality_test_prompts.go` | UPDATE; placeholders=$1; 顺序编号 | `IncrementQualityTestPromptUsage` (L103) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/quality_tests.go` | DDL, PG DDL/type, SQLite legacy | `ensureQualityTestSchema` (L125) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/quality_tests.go` | SELECT, INSERT, ON CONFLICT, RETURNING | `CreateQualityTestJob` (L166) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/quality_tests.go` | UPDATE; placeholders=$1 / $2; 重复/乱序需核对 | `ExpireQualityTests` (L200) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/quality_tests.go` | SELECT; placeholders=$1; 顺序编号 | `GetQualityTestJob` (L244) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/quality_tests.go` | SELECT; 通用 SQL/DB API | `ListQualityTests` (L248) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/quality_tests.go` | SELECT; 通用 SQL/DB API | `qualityTestFacets` (L289) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/quality_tests.go` | SELECT; placeholders=$1; 顺序编号 | `QualityTestStatus` (L342) | **B** | 待 placeholder 适配 | 已有同文件测试 |
| `database/quality_tests.go` | UPDATE; placeholders=$1 / $2 / $3 / $4; 顺序编号 | `SaveQualityTestProgress` (L348) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/quality_tests.go` | UPDATE; placeholders=$1 / $2 / $3 / $4 / $5 / $6; 重复/乱序需核对 | `FinishQualityTest` (L357) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/quality_tests.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `CancelQualityTest` (L366) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/response_cache_settings.go` | DB API; 通用 SQL/DB API | `GetResponseCacheSettings` (L122) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/response_cache_settings.go` | INSERT, UPDATE, TX, ON CONFLICT | `UpdateResponseCacheSettings` (L137) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有直接测试引用 |
| `database/response_cache_settings.go` | SELECT, UPDATE; 通用 SQL/DB API | `responseCacheSettingsSelectQuery` (L211) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/scheduler_outbox.go` | INSERT; placeholders=$1 / $2 / $3; 顺序编号 | `insertSchedulerOutboxEventTx` (L31) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/scheduler_outbox.go` | SELECT; 通用 SQL/DB API | `SchedulerOutboxHighWatermark` (L50) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/scheduler_outbox.go` | SELECT; placeholders=$1 / $2; 顺序编号 | `ListSchedulerOutboxEventsAfter` (L59) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/scheduler_outbox.go` | SELECT; 通用 SQL/DB API | `ListSchedulerOutboxEventsByIDs` (L89) | **A** | 无需方言改写 | 已有直接测试引用 |
| `database/scheduler_outbox.go` | SELECT, DELETE; placeholders=$1 / $2 / $3; 重复/乱序需核对 | `CleanupSchedulerOutboxThrough` (L129) | **B** | 待 placeholder 适配 | 已有直接测试引用 |
| `database/scheduler_outbox.go` | SELECT, INSERT, UPDATE, DELETE, ON CONFLICT, SQLite legacy | `installSQLiteSchedulerOutboxTriggers` (L164) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有同文件测试 |
| `database/scheduler_outbox.go` | INSERT, UPDATE, DELETE, ON CONFLICT, JSONB cast, JSONB operator, PG DDL/type | `installPostgresSchedulerOutboxTriggers` (L316) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |
| `database/sqlite.go` | SELECT, SQLite legacy | `withSQLiteWriteLock` (L63) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有同文件测试 |
| `database/sqlite.go` | TX, SQLite legacy | `withWriteTx` (L79) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/sqlite.go` | DB API, SQLite legacy | `configureSQLite` (L97) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有同文件测试 |
| `database/sqlite.go` | UPDATE, DDL, SQLite legacy | `migrateSQLite` (L111) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/sqlite.go` | DDL, SQLite legacy | `ensureSQLiteColumn` (L897) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有同文件测试 |
| `database/sqlite.go` | DB API, SQLite legacy | `sqliteTableColumns` (L909) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/sqlite.go` | SELECT, SQLite legacy | `getTrafficSnapshotSQLite` (L934) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有同文件测试 |
| `database/sqlite.go` | SELECT, SQLite legacy | `getChartAggregationSQLite` (L994) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有同文件测试 |
| `database/sqlite.go` | SELECT, SQLite legacy | `getAccountEventTrendSQLite` (L1071) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有同文件测试 |
| `database/sqlite.go` | SELECT, SQLite legacy | `getUsageStatsSQLite` (L1148) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有同文件测试 |
| `database/sqlite_schema_template.go` | DB API, SQLite legacy | `PrepareSQLiteSchemaTemplate` (L34) | **C** | 待移除 SQLite 语法并改为目标方言 | 已有直接测试引用 |
| `database/usage_snapshot.go` | UPDATE; placeholders=$1 / $2; 顺序编号 | `ClearCooldownIfReason` (L48) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/usage_snapshot.go` | UPDATE; placeholders=$1 / $2 / $3; 顺序编号 | `ClearCooldownIfReasonAndUntil` (L72) | **B** | 待 placeholder 适配 | 未发现直接测试 |
| `database/visible_channels_settings.go` | SELECT; 通用 SQL/DB API | `LoadVisibleChannelsConfig` (L59) | **A** | 无需方言改写 | 已有同文件测试 |
| `database/visible_channels_settings.go` | INSERT, UPDATE, ON CONFLICT | `SaveVisibleChannelsConfig` (L83) | **C** | 待 MySQL/PostgreSQL 方言兼容改写 | 已有同文件测试 |

## 6. S0.6 结论

- `database/` 已完成逐文件静态审计并形成 A/B/C 清单。
- B 类只记录 placeholder 适配要求；不在 S0.6 提前引入驱动包装或运行时转换机制。
- C 类明确记录 PostgreSQL/SQLite 专属 SQL 与 SQLite 遗留运行路径，后续按二次开发文档阶段顺序逐项迁移。
- MySQL Integration Test 属于后续阶段，本阶段不提前实现。
