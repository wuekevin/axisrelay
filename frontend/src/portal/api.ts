import type {
  AuthSessionResponse,
  PortalAPIKey,
  PortalSession,
  PortalSubscription,
  PortalUsageReport,
  PortalUser,
} from './types'

export class PortalAPIError extends Error {
  code: string
  status: number

  constructor(message: string, code = 'request_failed', status = 0) {
    super(message)
    this.name = 'PortalAPIError'
    this.code = code
    this.status = status
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  headers.set('X-Requested-With', 'AxisRelayPortal')
  const response = await fetch(path, {
    ...init,
    headers,
    credentials: 'same-origin',
  })
  if (response.status === 204) return undefined as T
  const payload = await response.json().catch(() => ({}))
  if (!response.ok) {
    const error = payload?.error ?? {}
    throw new PortalAPIError(error.message || '请求失败，请稍后重试', error.code, response.status)
  }
  return payload as T
}

export const portalAPI = {
  session: () => request<AuthSessionResponse>('/api/auth/session'),
  register: (payload: { email: string; password: string; display_name: string; accept_terms: boolean }) =>
    request<AuthSessionResponse>('/api/auth/register', { method: 'POST', body: JSON.stringify(payload) }),
  login: (payload: { email: string; password: string; remember: boolean }) =>
    request<AuthSessionResponse>('/api/auth/login', { method: 'POST', body: JSON.stringify(payload) }),
  logout: () => request<void>('/api/auth/logout', { method: 'POST' }),
  profile: () => request<{ user: PortalUser; subscription: PortalSubscription }>('/api/user/profile'),
  updateProfile: (payload: Pick<PortalUser, 'display_name' | 'phone' | 'locale' | 'timezone'>) =>
    request<{ user: PortalUser }>('/api/user/profile', { method: 'PATCH', body: JSON.stringify(payload) }),
  changePassword: (payload: { current_password: string; new_password: string }) =>
    request<{ message: string }>('/api/user/password', { method: 'PUT', body: JSON.stringify(payload) }),
  sessions: () => request<{ sessions: PortalSession[] }>('/api/user/sessions'),
  revokeSession: (id: number) => request<void>(`/api/user/sessions/${id}`, { method: 'DELETE' }),
  revokeOtherSessions: () => request<{ message: string }>('/api/user/sessions/revoke-others', { method: 'POST' }),
  apiKeys: () => request<{ api_keys: PortalAPIKey[]; limit: number }>('/api/user/api-keys'),
  createAPIKey: (name: string) => request<{ api_key: PortalAPIKey; secret: string }>('/api/user/api-keys', {
    method: 'POST', body: JSON.stringify({ name }),
  }),
  renameAPIKey: (id: number, name: string) => request<{ message: string }>(`/api/user/api-keys/${id}`, {
    method: 'PATCH', body: JSON.stringify({ name }),
  }),
  revokeAPIKey: (id: number) => request<void>(`/api/user/api-keys/${id}`, { method: 'DELETE' }),
  usage: (id: number, days: number) => request<{ range_days: number; usage: PortalUsageReport }>(
    `/api/user/api-keys/${id}/usage?days=${days}`,
  ),
  subscription: () => request<{ subscription: PortalSubscription }>('/api/user/subscription'),
}
