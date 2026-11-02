import type { FormEvent, PropsWithChildren } from 'react'
import { useCallback, useEffect, useState } from 'react'
import { KeyRound, LoaderCircle, ShieldCheck, UserRound } from 'lucide-react'
import { ADMIN_AUTH_REQUIRED_EVENT } from '../api'
import { DEFAULT_SITE_LOGO, useBranding } from '../branding'

type AuthState = 'checking' | 'setup' | 'login' | 'change-password' | 'authenticated'
interface AuthStatus { authenticated?: boolean; needs_setup?: boolean; user?: { must_change_password?: boolean } }

export default function AuthGate({ children }: PropsWithChildren) {
  const { siteName: brandedSiteName, siteLogo } = useBranding()
  const [state, setState] = useState<AuthState>('checking')
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [remember, setRemember] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const refresh = useCallback(async () => {
    setState('checking')
    try {
      const response = await fetch('/api/admin/auth/status', { cache: 'no-store' })
      if (!response.ok) throw new Error('暂时无法读取登录状态')
      const result = (await response.json()) as AuthStatus
      setState(result.authenticated ? (result.user?.must_change_password ? 'change-password' : 'authenticated') : result.needs_setup ? 'setup' : 'login')
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : '后台服务暂不可用')
      setState('login')
    }
  }, [])

  useEffect(() => {
    void refresh()
    const requireLogin = () => { setPassword(''); setState('login') }
    window.addEventListener(ADMIN_AUTH_REQUIRED_EVENT, requireLogin)
    return () => window.removeEventListener(ADMIN_AUTH_REQUIRED_EVENT, requireLogin)
  }, [refresh])

  const submit = async (event: FormEvent) => {
    event.preventDefault(); setError('')
    if ((state === 'setup' || state === 'change-password') && password !== confirmPassword) { setError('两次输入的密码不一致'); return }
    if ((state === 'setup' || state === 'change-password') && password.length < 12) { setError('管理员密码至少 12 位'); return }
    if (state === 'login' && password.length === 0) { setError('请输入管理员密码'); return }
    setSubmitting(true)
    try {
      const endpoint = state === 'setup' ? '/api/admin/auth/setup' : state === 'change-password' ? '/api/admin/auth/password' : '/api/admin/auth/login'
      const response = await fetch(endpoint, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(state === 'setup' ? { username, password, confirm_password: confirmPassword } : state === 'change-password' ? { password, confirm_password: confirmPassword } : { username, password, remember }) })
      const result = await response.json().catch(() => ({})) as { error?: string; must_change_password?: boolean }
      if (!response.ok) throw new Error(result.error || '登录失败，请稍后再试')
      setPassword(''); setConfirmPassword(''); setState(result.must_change_password ? 'change-password' : 'authenticated')
    } catch (cause) { setError(cause instanceof Error ? cause.message : '登录失败，请稍后再试') } finally { setSubmitting(false) }
  }

  if (state === 'authenticated') return <>{children}</>
  const siteName = brandedSiteName || 'AxisRelay'
  const logo = siteLogo || DEFAULT_SITE_LOGO
  const isSetup = state === 'setup'
  const isPasswordChange = state === 'change-password'
  return (
    <main className="relative grid min-h-screen place-items-center overflow-hidden bg-background px-4 py-10 text-foreground">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_20%_10%,color-mix(in_oklab,var(--color-primary)_16%,transparent),transparent_42%),radial-gradient(circle_at_85%_85%,color-mix(in_oklab,var(--color-primary)_10%,transparent),transparent_38%)]" />
      <section className="relative w-full max-w-md rounded-3xl border border-border/70 bg-card/95 p-7 shadow-2xl shadow-black/10 backdrop-blur sm:p-9">
        <div className="mb-8 flex items-center gap-3"><img src={logo} alt="" className="size-11 rounded-xl object-contain" /><div><p className="text-lg font-semibold">{siteName}</p><p className="text-sm text-muted-foreground">安全管理控制台</p></div></div>
        {state === 'checking' ? <div className="flex min-h-48 flex-col items-center justify-center gap-3 text-muted-foreground"><LoaderCircle className="size-7 animate-spin" /><span>正在验证安全会话…</span></div> : (
          <form onSubmit={submit} className="space-y-5">
            <div><div className="mb-3 inline-flex size-11 items-center justify-center rounded-2xl bg-primary/10 text-primary">{isSetup ? <ShieldCheck /> : <KeyRound />}</div><h1 className="text-2xl font-semibold tracking-tight">{isSetup ? '创建管理员账号' : isPasswordChange ? '修改初始密码' : '管理员登录'}</h1><p className="mt-2 text-sm leading-6 text-muted-foreground">{isSetup ? '首次使用请创建独立管理员账号。密码仅以安全哈希保存。' : isPasswordChange ? '旧版管理密钥已迁移为初始密码，请设置新的独立管理员密码。' : '请输入管理员账号和密码进入控制台。'}</p></div>
            {!isPasswordChange && <label className="block space-y-2"><span className="text-sm font-medium">管理员账号</span><div className="relative"><UserRound className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" /><input autoComplete="username" value={username} onChange={(e) => setUsername(e.target.value)} className="h-11 w-full rounded-xl border border-input bg-background pl-10 pr-3 outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20" /></div></label>}
            <label className="block space-y-2"><span className="text-sm font-medium">密码</span><input type="password" maxLength={72} autoComplete={isSetup ? 'new-password' : 'current-password'} value={password} onChange={(e) => setPassword(e.target.value)} className="h-11 w-full rounded-xl border border-input bg-background px-3 outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20" placeholder="至少 12 位" /></label>
            {(isSetup || isPasswordChange) && <label className="block space-y-2"><span className="text-sm font-medium">确认密码</span><input type="password" maxLength={72} autoComplete="new-password" value={confirmPassword} onChange={(e) => setConfirmPassword(e.target.value)} className="h-11 w-full rounded-xl border border-input bg-background px-3 outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20" /></label>}
            {!isSetup && <label className="flex items-center gap-2 text-sm text-muted-foreground"><input type="checkbox" checked={remember} onChange={(e) => setRemember(e.target.checked)} className="size-4 rounded border-input" />在此设备保持登录</label>}
            {error && <p role="alert" className="rounded-xl border border-destructive/20 bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</p>}
            <button disabled={submitting} className="flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-primary font-medium text-primary-foreground shadow-lg shadow-primary/20 transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60">{submitting && <LoaderCircle className="size-4 animate-spin" />}{isSetup ? '创建并进入后台' : isPasswordChange ? '保存新密码' : '登录'}</button>
          </form>
        )}
      </section>
    </main>
  )
}
