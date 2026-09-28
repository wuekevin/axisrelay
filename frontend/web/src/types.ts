export type ISODateString = string
export type ToastType = 'success' | 'error' | 'warning' | 'info'
export type UpstreamChannel = 'codex' | 'grok' | 'antigravity' | 'claude'

export interface ToastState {
  msg: string
  type: ToastType
}

export interface MessageResponse {
  message: string
  warning?: string
}

export interface SiteBranding {
  site_name: string
  site_logo: string
  background_image: string
  background_opacity: number
  background_blur: number
  background_glass_opacity: number
  background_glass_blur: number
}

export interface AccountPortalAuthURLResponse {
  auth_url: string
  session_id: string
}

export interface AccountPortalSubmitResponse {
  message: string
}

export interface APIKeyScopeLimit {
  scope_type: 'group' | 'account'
  scope_id: number
  on_exhausted?: 'skip' | 'reject'
  cost_5h?: number
  cost_1d?: number
  cost_7d?: number
  cost_30d?: number
  token_5h?: number
  token_1d?: number
  token_7d?: number
  token_30d?: number
  requests_1d?: number
  max_concurrency?: number
  quota_cost?: number
  quota_tokens?: number
  quota_requests?: number
}

export interface APIKeyModelRequestLimit {
  id?: string
  model: string
  window: 'week'
  max_requests: number
  timezone: string
  reset_weekday: number
  reset_time: string
}

export interface APIKeyModelRequestUsage {
  rule_id: string
  model: string
  window: 'week'
  limit: number
  used: number
  remaining: number
  window_start: ISODateString
  reset_at: ISODateString
  timezone: string
}

export interface APIKeyLimits {
  model_allow?: string[]
  model_deny?: string[]
  plan_allow?: string[]
  no_affinity_group_ids?: number[]
  rpm?: number
  rpd?: number
  max_concurrency?: number
  cost_limit_5h?: number
  cost_limit_7d?: number
  cost_limit_30d?: number
  cost_limit_daily?: number
  token_limit_5h?: number
  token_limit_7d?: number
  token_limit_30d?: number
  token_limit_daily?: number
  disable_image_generation?: boolean
  image_generation_policy?: 'allow' | 'strip' | 'block'
  upstream_channel?: UpstreamChannel
  allow_live?: boolean
  scope_limits?: APIKeyScopeLimit[]
  model_request_limits?: APIKeyModelRequestLimit[]
}

export interface PublicAPIKeyUsageKey {
  name: string
  key: string
  quota_limit: number
  quota_used: number
  total_used: number
  reset_count: number
  last_reset_at?: ISODateString | null
  expires_at?: ISODateString | null
  limits: APIKeyLimits
  status: 'active' | 'expired' | 'quota_exhausted'
  created_at: ISODateString
}

export interface PublicAPIKeyUsageRange {
  name: 'today' | '7d' | '30d' | 'all' | string
  start?: ISODateString | null
  end: ISODateString
}

export interface PublicAPIKeyWindowUsage {
  requests: number
  tokens: number
  user_billed: number
  oldest_at?: ISODateString
  window_kind: 'fixed' | 'sliding'
  reset_at?: ISODateString
  decay_at?: ISODateString
}

export interface PublicAPIKeyUsageWindows {
  today: PublicAPIKeyWindowUsage
  last_5h: PublicAPIKeyWindowUsage
  last_7d: PublicAPIKeyWindowUsage
  last_30d: PublicAPIKeyWindowUsage
}

export interface PublicAPIKeyUsageSummary {
  requests: number
  tokens: number
  input_tokens: number
  output_tokens: number
  cached_tokens: number
  error_count: number
  user_billed: number
  avg_duration_ms: number
  avg_first_token_ms: number
  rpm: number
  tpm: number
}

export interface PublicAPIKeyUsageBreakdown {
  name: string
  requests: number
  tokens: number
  input_tokens: number
  output_tokens: number
  cached_tokens: number
  error_count: number
  user_billed: number
}

export interface PublicAPIKeyUsageLog {
  user_billing_mode?: '' | 'token' | 'per_image'
  image_unit_price?: number
  billed_image_count?: number
  id: number
  endpoint: string
  model: string
  effective_model: string
  status_code: number
  duration_ms: number
  first_token_ms: number
  input_tokens: number
  output_tokens: number
  cached_tokens: number
  total_tokens: number
  user_billed: number
  input_cost: number
  output_cost: number
  cache_read_cost: number
  total_cost: number
  input_price_per_mtoken: number
  output_price_per_mtoken: number
  cache_read_price_per_mtoken: number
  rate_multiplier: number
  long_context: boolean
  service_tier: string
  stream: boolean
  compact: boolean
  has_compaction_history: boolean
  ultra?: boolean
  via_websocket: boolean
  upstream_error_kind: string
  created_at: ISODateString
}

export interface PublicAPIKeyUsageReport {
  summary: PublicAPIKeyUsageSummary
  windows: PublicAPIKeyUsageWindows
  models: PublicAPIKeyUsageBreakdown[]
  endpoints: PublicAPIKeyUsageBreakdown[]
  recent_logs: PublicAPIKeyUsageLog[]
  recent_logs_total: number
  recent_logs_page: number
  recent_logs_page_size: number
}

export interface PublicAPIKeyUsageResponse {
  key: PublicAPIKeyUsageKey
  range: PublicAPIKeyUsageRange
  usage: PublicAPIKeyUsageReport
  model_request_usage?: APIKeyModelRequestUsage[]
}

export interface ImageStudioQuota {
  image_pricing?: Record<string, { user_billing_mode: 'token' | 'per_image'; image_unit_price?: number }>
  quota_limit: number
  quota_used: number
  quota_remaining: number | null
  expires_at: ISODateString | null
  status: 'active' | 'expired' | 'quota_exhausted'
  refresh_after_seconds: number
}

export interface ImageAsset {
  id: number
  job_id: number
  template_id: number
  filename: string
  proxy_url?: string
  thumbnail_url?: string
  mime_type: string
  bytes: number
  width: number
  height: number
  model: string
  requested_size: string
  actual_size: string
  quality: string
  output_format: string
  revised_prompt: string
  created_at: ISODateString
  cache_b64_json?: string
}

export interface ImageGenerationJob {
  id: number
  status: 'queued' | 'running' | 'succeeded' | 'failed' | string
  prompt: string
  params_json: string
  api_key_id: number
  api_key_name: string
  api_key_masked: string
  error_message: string
  warning?: string
  duration_ms: number
  created_at: ISODateString
  started_at?: ISODateString
  completed_at?: ISODateString
  assets?: ImageAsset[]
}

export interface ImageJobResponse {
  job: ImageGenerationJob
}

export interface ImageJobsResponse {
  jobs: ImageGenerationJob[]
  total: number
}

export interface ImageAssetsResponse {
  assets: ImageAsset[]
  total: number
}

export interface CreateImageJobPayload {
  prompt: string
  model?: string
  size?: string
  quality?: string
  output_format?: string
  background?: string
  style?: string
  upscale?: string
  strict_size?: boolean
  upscale_fit?: 'pad' | 'cover'
  api_key_id?: number
  template_id?: number
  input_images?: string[]
}
