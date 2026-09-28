export type PortalUser = {
  id: string
  email: string
  status: string
  email_verified_at?: string
  last_login_at?: string
  created_at: string
  display_name: string
  avatar_url: string
  phone: string
  locale: 'zh-CN' | 'zh-TW' | 'en'
  timezone: string
}

export type PortalSubscription = {
  code: string
  name: string
  status: string
  billing_period: string
  ends_at?: string
}

export type PortalSession = {
  id: number
  client_ip: string
  user_agent: string
  created_at: string
  last_seen_at?: string
  expires_at: string
  current: boolean
}

export type PortalAPIKey = {
  id: number
  name: string
  prefix: string
  status: string
  enabled: boolean
  quota_limit: number
  quota_used: number
  last_used_at?: string
  expires_at?: string
  created_at: string
}

export type PortalUsageSummary = {
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

export type PortalUsageBreakdown = {
  name: string
  requests: number
  tokens: number
  input_tokens: number
  output_tokens: number
  cached_tokens: number
  error_count: number
  user_billed: number
}

export type PortalUsageLog = {
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
  created_at: string
}

export type PortalUsageReport = {
  summary: PortalUsageSummary
  models: PortalUsageBreakdown[]
  endpoints: PortalUsageBreakdown[]
  recent_logs: PortalUsageLog[]
  recent_logs_total: number
}

export type AuthSessionResponse = {
  authenticated: boolean
  user?: PortalUser
}
