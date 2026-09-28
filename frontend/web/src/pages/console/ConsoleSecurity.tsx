import { Globe2, KeyRound, Laptop, Loader2, LockKeyhole, LogOut, ShieldCheck, Smartphone } from 'lucide-react'
import { useEffect, useState, type FormEvent } from 'react'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { portalAPI, PortalAPIError } from '../../portal/api'
import { formatDate, shortUserAgent } from '../../portal/format'
import type { PortalSession } from '../../portal/types'
import { useUserAuth } from '../../providers/AppProviders'

export default function ConsoleSecurity() {
  const auth = useUserAuth()
  const [sessions, setSessions] = useState<PortalSession[]>([])
  const [loading, setLoading] = useState(true)
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  const load = async () => {
    setLoading(true)
    try { setSessions((await portalAPI.sessions()).sessions) }
    catch (reason) { setError(reason instanceof Error ? reason.message : '读取失败') }
    finally { setLoading(false) }
  }
  useEffect(() => { void load() }, [])

  const changePassword = async (event: FormEvent) => {
    event.preventDefault()
    setMessage('')
    setError('')
    if (newPassword !== confirmPassword) { setError('两次输入的新密码不一致'); return }
    setSaving(true)
    try {
      const response = await portalAPI.changePassword({ current_password: currentPassword, new_password: newPassword })
      setMessage(response.message)
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
      await load()
    } catch (reason) { setError(reason instanceof PortalAPIError ? reason.message : '密码更新失败') }
    finally { setSaving(false) }
  }
  const revoke = async (session: PortalSession) => {
    await portalAPI.revokeSession(session.id)
    if (session.current) { await auth.refresh(); return }
    await load()
  }
  const revokeOthers = async () => {
    try { const response = await portalAPI.revokeOtherSessions(); setMessage(response.message); await load() }
    catch (reason) { setError(reason instanceof Error ? reason.message : '操作失败') }
  }

  return <div className="mx-auto max-w-5xl space-y-6">
    <div><h2 className="text-2xl font-bold tracking-tight">账号安全</h2><p className="mt-1 text-sm text-muted-foreground">更新密码并管理当前账号的全部登录设备。</p></div>
    <div className="grid gap-5 lg:grid-cols-2">
      <form onSubmit={changePassword} className="rounded-2xl border border-border bg-card p-6 shadow-sm">
        <div className="flex items-center gap-3"><span className="flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary"><LockKeyhole className="size-5" /></span><div><h3 className="font-semibold">修改密码</h3><p className="text-xs text-muted-foreground">保存后其他设备会自动退出</p></div></div>
        <div className="mt-6 space-y-4"><Field label="当前密码"><Input type="password" value={currentPassword} onChange={(event) => setCurrentPassword(event.target.value)} autoComplete="current-password" required /></Field><Field label="新密码"><Input type="password" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} autoComplete="new-password" minLength={10} required placeholder="至少 10 位，含大小写字母和数字" /></Field><Field label="确认新密码"><Input type="password" value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} autoComplete="new-password" minLength={10} required /></Field></div>
        {message && <div className="mt-5 rounded-lg border border-emerald-500/20 bg-emerald-500/8 px-3 py-2 text-sm text-emerald-600">{message}</div>}
        {error && <div className="mt-5 rounded-lg border border-red-500/20 bg-red-500/8 px-3 py-2 text-sm text-red-600">{error}</div>}
        <Button type="submit" className="mt-5 w-full" disabled={saving}>{saving ? <Loader2 className="animate-spin" /> : <KeyRound />}{saving ? '更新中…' : '更新密码'}</Button>
      </form>
      <div className="rounded-2xl border border-border bg-card p-6 shadow-sm">
        <div className="flex items-center gap-3"><span className="flex size-10 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-600"><ShieldCheck className="size-5" /></span><div><h3 className="font-semibold">安全状态</h3><p className="text-xs text-muted-foreground">关键保护已启用</p></div></div>
        <div className="mt-6 space-y-3 text-sm"><SecurityLine icon={LockKeyhole} label="密码保护" value="已启用" /><SecurityLine icon={Globe2} label="异常访问防护" value="已启用" /><SecurityLine icon={ShieldCheck} label="邮箱验证" value="已完成" /></div>
        <div className="mt-6 rounded-xl border border-border bg-muted/40 p-4 text-xs leading-6 text-muted-foreground">如发现陌生设备，请立即退出该设备并更新密码。修改密码后，其他设备会自动退出。</div>
      </div>
    </div>
    <section className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
      <div className="flex flex-col justify-between gap-4 border-b border-border p-5 sm:flex-row sm:items-center"><div><h3 className="font-semibold">登录设备</h3><p className="mt-1 text-xs text-muted-foreground">当前有效会话会显示在这里。</p></div><Button variant="outline" size="sm" onClick={() => void revokeOthers()} disabled={sessions.filter((item) => !item.current).length === 0}><LogOut />退出其他设备</Button></div>
      {loading ? <div className="flex h-32 items-center justify-center text-muted-foreground"><Loader2 className="mr-2 size-4 animate-spin" />正在读取</div> : sessions.length === 0 ? <div className="p-8 text-center text-sm text-muted-foreground">暂无登录设备</div> : <div className="divide-y divide-border">{sessions.map((session) => <div key={session.id} className="flex flex-col gap-4 p-5 sm:flex-row sm:items-center"><span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-muted">{session.user_agent.toLowerCase().includes('mobile') ? <Smartphone className="size-5" /> : <Laptop className="size-5" />}</span><div className="min-w-0 flex-1"><div className="flex flex-wrap items-center gap-2"><span className="font-medium">{shortUserAgent(session.user_agent)}</span>{session.current && <span className="rounded-full bg-emerald-500/10 px-2 py-0.5 text-[11px] font-medium text-emerald-600">当前设备</span>}</div><div className="mt-1 text-xs text-muted-foreground">IP {session.client_ip || '未知'} · 最近活动 {formatDate(session.last_seen_at || session.created_at)} · 到期 {formatDate(session.expires_at)}</div></div><Button variant="ghost" size="sm" className="text-destructive" onClick={() => void revoke(session)}>{session.current ? '退出当前设备' : '退出设备'}</Button></div>)}</div>}
    </section>
  </div>
}

function Field({ label, children }: { label: string; children: React.ReactNode }) { return <label className="block"><span className="mb-1.5 block text-sm font-medium">{label}</span>{children}</label> }
function SecurityLine({ icon: Icon, label, value }: { icon: typeof LockKeyhole; label: string; value: string }) { return <div className="flex items-center gap-3 rounded-lg border border-border p-3"><Icon className="size-4 text-primary" /><span className="flex-1 text-muted-foreground">{label}</span><span className="font-medium">{value}</span></div> }
