import type {
  AccountPortalAuthURLResponse,
  AccountPortalSubmitResponse,
  CreateImageJobPayload,
  ImageAssetsResponse,
  ImageJobResponse,
  ImageJobsResponse,
  ImageStudioQuota,
  MessageResponse,
  PublicAPIKeyUsageResponse,
  SiteBranding,
} from './types'

type HTTPError = Error & { status?: number }

function errorMessage(body: string, status: number): string {
  if (!body.trim()) return `HTTP ${status}`
  try {
    const parsed = JSON.parse(body) as { error?: { message?: string } | string; message?: string }
    if (typeof parsed.error === 'object' && parsed.error?.message) return parsed.error.message
    if (typeof parsed.error === 'string' && parsed.error.trim()) return parsed.error
    if (typeof parsed.message === 'string' && parsed.message.trim()) return parsed.message
  } catch {
    // Keep the server response when it is not JSON.
  }
  return body
}

async function publicJSON<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  if (options.body !== undefined && options.body !== null && !(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  const response = await fetch(path, { ...options, cache: options.cache ?? 'no-store', headers })
  if (!response.ok) {
    const error = new Error(errorMessage(await response.text(), response.status)) as HTTPError
    error.status = response.status
    throw error
  }
  if (response.status === 204) return undefined as T
  const text = await response.text()
  return (text ? JSON.parse(text) : undefined) as T
}

function authorizedJSON<T>(base: string, path: string, apiKey: string, options: RequestInit = {}) {
  const headers = new Headers(options.headers)
  headers.set('Authorization', `Bearer ${apiKey}`)
  return publicJSON<T>(base + path, { ...options, headers })
}

async function authorizedBlob(path: string, apiKey: string): Promise<Blob> {
  const response = await fetch('/api/image-studio' + path, {
    cache: 'no-store',
    headers: { Authorization: `Bearer ${apiKey}` },
  })
  if (!response.ok) throw new Error(errorMessage(await response.text(), response.status))
  return response.blob()
}

export const api = {
  getBranding: () => publicJSON<SiteBranding>('/api/branding'),
  generateAccountPortalAuthURL: (data: { contact_email: string }) =>
    publicJSON<AccountPortalAuthURLResponse>('/api/account-portal/generate-auth-url', { method: 'POST', body: JSON.stringify(data) }),
  submitAccountPortalCode: (data: { session_id: string; code: string; state: string }) =>
    publicJSON<AccountPortalSubmitResponse>('/api/account-portal/submit-code', { method: 'POST', body: JSON.stringify(data) }),
  getPublicAPIKeyUsage: (apiKey: string, range = '30d', params: { page?: number; pageSize?: number } = {}) => {
    const search = new URLSearchParams({ range })
    if (params.page) search.set('page', String(params.page))
    if (params.pageSize) search.set('page_size', String(params.pageSize))
    return authorizedJSON<PublicAPIKeyUsageResponse>('/api/key-usage', `/summary?${search}`, apiKey)
  },
  getPortalImageQuota: (apiKey: string) => authorizedJSON<ImageStudioQuota>('/api/image-studio', '/quota', apiKey),
  createPortalImageJob: (apiKey: string, data: CreateImageJobPayload) =>
    authorizedJSON<ImageJobResponse>('/api/image-studio', '/jobs', apiKey, { method: 'POST', body: JSON.stringify(data) }),
  createPortalImageEditJob: (apiKey: string, data: CreateImageJobPayload) =>
    authorizedJSON<ImageJobResponse>('/api/image-studio', '/edit-jobs', apiKey, { method: 'POST', body: JSON.stringify(data) }),
  getPortalImageJobs: (apiKey: string, params: { page?: number; pageSize?: number } = {}) => {
    const search = new URLSearchParams()
    if (params.page) search.set('page', String(params.page))
    if (params.pageSize) search.set('page_size', String(params.pageSize))
    return authorizedJSON<ImageJobsResponse>('/api/image-studio', `/jobs?${search}`, apiKey)
  },
  getPortalImageJob: (apiKey: string, id: number, params: { includeCache?: boolean } = {}) =>
    authorizedJSON<ImageJobResponse>('/api/image-studio', `/jobs/${id}${params.includeCache ? '?include_cache=1' : ''}`, apiKey),
  deletePortalImageJob: (apiKey: string, id: number) =>
    authorizedJSON<MessageResponse>('/api/image-studio', `/jobs/${id}`, apiKey, { method: 'DELETE' }),
  getPortalImageAssets: (apiKey: string, params: { page?: number; pageSize?: number } = {}) => {
    const search = new URLSearchParams()
    if (params.page) search.set('page', String(params.page))
    if (params.pageSize) search.set('page_size', String(params.pageSize))
    return authorizedJSON<ImageAssetsResponse>('/api/image-studio', `/assets?${search}`, apiKey)
  },
  getPortalImageAssetFile: (apiKey: string, id: number, download = false, thumbKB = 0) => {
    const search = new URLSearchParams()
    if (download) search.set('download', '1')
    if (thumbKB > 0) search.set('thumb_kb', String(thumbKB))
    const query = search.toString()
    return authorizedBlob(`/assets/${id}/file${query ? `?${query}` : ''}`, apiKey)
  },
  deletePortalImageAsset: (apiKey: string, id: number) =>
    authorizedJSON<MessageResponse>('/api/image-studio', `/assets/${id}`, apiKey, { method: 'DELETE' }),
}
