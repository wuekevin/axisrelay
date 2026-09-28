import { Check, Copy, KeyRound, Loader2, MoreHorizontal, Pencil, Plus, ShieldAlert, Trash2, X } from 'lucide-react'
import { useEffect, useState, type FormEvent } from 'react'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { portalAPI, PortalAPIError } from '../../portal/api'
import { formatDate, formatUSD } from '../../portal/format'
import type { PortalAPIKey } from '../../portal/types'

export default function ConsoleAPIKeys() {
  const [keys, setKeys] = useState<PortalAPIKey[]>([])
  const [limit, setLimit] = useState(10)
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [name, setName] = useState('')
  const [secret, setSecret] = useState('')
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState('')
  const [editing, setEditing] = useState<PortalAPIKey | null>(null)
  const [revokeTarget, setRevokeTarget] = useState<PortalAPIKey | null>(null)

  const load = async () => {
    setLoading(true)
    try { const response = await portalAPI.apiKeys(); setKeys(response.api_keys); setLimit(response.limit) }
    catch (reason) { setError(reason instanceof Error ? reason.message : '读取失败') }
    finally { setLoading(false) }
  }
  useEffect(() => { void load() }, [])

  const create = async (event: FormEvent) => {
    event.preventDefault(); setError(''); setCreating(true)
    try { const response = await portalAPI.createAPIKey(name); setSecret(response.secret); setName(''); await load() }
    catch (reason) { setError(reason instanceof PortalAPIError ? reason.message : '创建失败') }
    finally { setCreating(false) }
  }

  const rename = async (event: FormEvent) => {
    event.preventDefault(); if (!editing) return
    try { await portalAPI.renameAPIKey(editing.id, editing.name); setEditing(null); await load() }
    catch (reason) { setError(reason instanceof Error ? reason.message : '更新失败') }
  }

  const revoke = async () => {
    if (!revokeTarget) return
    try { await portalAPI.revokeAPIKey(revokeTarget.id); setRevokeTarget(null); await load() }
    catch (reason) { setError(reason instanceof Error ? reason.message : '撤销失败') }
  }

  return <div className="space-y-6"><div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end"><div><h2 className="text-2xl font-bold tracking-tight">API Key</h2><p className="mt-1 text-sm text-muted-foreground">创建和撤销真实可调用 `/v1/*` 的访问凭据。</p></div><span className="text-sm text-muted-foreground">{keys.length} / {limit}</span></div>
    <form onSubmit={create} className="flex flex-col gap-3 rounded-2xl border border-border bg-card p-5 shadow-sm sm:flex-row"><div className="flex-1"><label className="mb-1.5 block text-sm font-medium">新 Key 名称</label><Input value={name} onChange={(event) => setName(event.target.value)} placeholder="例如：Production API" minLength={2} maxLength={64} required className="h-10" /></div><Button type="submit" className="mt-auto h-10" disabled={creating || keys.length >= limit}>{creating ? <Loader2 className="animate-spin" /> : <Plus />}创建 Key</Button></form>
    {error && <div className="rounded-xl border border-red-500/20 bg-red-500/8 p-4 text-sm text-red-600 dark:text-red-400">{error}</div>}
    <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">{loading ? <div className="flex h-48 items-center justify-center text-muted-foreground"><Loader2 className="mr-2 size-4 animate-spin" />正在读取</div> : keys.length === 0 ? <div className="flex h-56 flex-col items-center justify-center px-6 text-center"><span className="flex size-12 items-center justify-center rounded-xl bg-muted"><KeyRound className="size-5 text-muted-foreground" /></span><h3 className="mt-4 font-semibold">还没有 API Key</h3><p className="mt-1 text-sm text-muted-foreground">在上方输入名称创建第一把 Key。</p></div> : <div className="divide-y divide-border">{keys.map((item) => <div key={item.id} className="grid gap-4 p-5 sm:grid-cols-[1fr_auto] sm:items-center"><div className="min-w-0"><div className="flex items-center gap-2"><h3 className="truncate font-semibold">{item.name}</h3><span className="rounded-full bg-emerald-500/10 px-2 py-0.5 text-[11px] font-medium text-emerald-600">有效</span></div><div className="mt-2 flex flex-wrap gap-x-5 gap-y-1 text-xs text-muted-foreground"><span className="font-mono">{item.prefix}••••••</span><span>创建于 {formatDate(item.created_at)}</span><span>最近使用 {formatDate(item.last_used_at)}</span></div><div className="mt-3 flex items-center gap-3 text-xs"><span>额度 {formatUSD(item.quota_used)} / {formatUSD(item.quota_limit)}</span><span className="h-1.5 w-24 overflow-hidden rounded-full bg-muted"><span className="block h-full bg-primary" style={{ width: `${Math.min(100, item.quota_limit > 0 ? item.quota_used / item.quota_limit * 100 : 0)}%` }} /></span></div></div><div className="flex gap-2"><Button variant="outline" size="sm" onClick={() => setEditing({ ...item })}><Pencil />重命名</Button><Button variant="ghost" size="icon-sm" aria-label="撤销 Key" className="text-destructive" onClick={() => setRevokeTarget(item)}><Trash2 /></Button></div></div>)}</div>}</div>
    {secret && <Overlay onClose={() => setSecret('')}><div className="flex size-11 items-center justify-center rounded-xl bg-amber-500/10 text-amber-600"><ShieldAlert /></div><h3 className="mt-4 text-xl font-bold">请立即保存 API Key</h3><p className="mt-2 text-sm leading-6 text-muted-foreground">出于安全原因，这个完整 Key 只显示一次。关闭后无法再次查看。</p><div className="mt-5 flex items-center gap-2 rounded-xl border border-border bg-muted/60 p-3"><code className="min-w-0 flex-1 break-all text-xs">{secret}</code><Button size="icon-sm" variant="outline" onClick={() => { void navigator.clipboard.writeText(secret); setCopied(true) }}>{copied ? <Check /> : <Copy />}</Button></div><Button className="mt-5 w-full" onClick={() => setSecret('')}>我已安全保存</Button></Overlay>}
    {editing && <Overlay onClose={() => setEditing(null)}><h3 className="text-xl font-bold">重命名 API Key</h3><form onSubmit={rename} className="mt-5 space-y-4"><Input value={editing.name} onChange={(event) => setEditing({ ...editing, name: event.target.value })} minLength={2} maxLength={64} required autoFocus /><div className="flex justify-end gap-2"><Button type="button" variant="outline" onClick={() => setEditing(null)}>取消</Button><Button type="submit">保存</Button></div></form></Overlay>}
    {revokeTarget && <Overlay onClose={() => setRevokeTarget(null)}><div className="flex size-11 items-center justify-center rounded-xl bg-red-500/10 text-red-600"><Trash2 /></div><h3 className="mt-4 text-xl font-bold">撤销“{revokeTarget.name}”？</h3><p className="mt-2 text-sm leading-6 text-muted-foreground">撤销立即生效，所有使用这个 Key 的应用都会停止工作。该操作不能恢复。</p><div className="mt-6 flex justify-end gap-2"><Button variant="outline" onClick={() => setRevokeTarget(null)}>取消</Button><Button variant="destructive" onClick={() => void revoke()}>确认撤销</Button></div></Overlay>}
  </div>
}

function Overlay({ children, onClose }: { children: React.ReactNode; onClose: () => void }) { return <div className="fixed inset-0 z-[70] flex items-center justify-center p-4"><button aria-label="关闭" className="absolute inset-0 bg-black/55 backdrop-blur-sm" onClick={onClose} /><div className="relative w-full max-w-md rounded-2xl border border-border bg-card p-6 shadow-2xl"><button className="absolute right-4 top-4 text-muted-foreground hover:text-foreground" onClick={onClose}><X className="size-4" /></button>{children}</div></div> }
