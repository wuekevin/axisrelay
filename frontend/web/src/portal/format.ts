export function formatNumber(value?: number) {
  return new Intl.NumberFormat('zh-CN', { notation: (value ?? 0) >= 100000 ? 'compact' : 'standard', maximumFractionDigits: 1 }).format(value ?? 0)
}

export function formatUSD(value?: number) {
  return new Intl.NumberFormat('zh-CN', { style: 'currency', currency: 'USD', minimumFractionDigits: 2, maximumFractionDigits: 4 }).format(value ?? 0)
}

export function formatDate(value?: string) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleString('zh-CN', { dateStyle: 'medium', timeStyle: 'short' })
}

export function shortUserAgent(value: string) {
  if (!value) return '未知设备'
  if (value.includes('Edg/')) return 'Microsoft Edge'
  if (value.includes('Chrome/')) return 'Google Chrome'
  if (value.includes('Safari/') && !value.includes('Chrome/')) return 'Safari'
  if (value.includes('Firefox/')) return 'Firefox'
  return value.length > 56 ? `${value.slice(0, 56)}…` : value
}
