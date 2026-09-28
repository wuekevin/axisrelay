import { ArrowRight, CheckCircle2, Eye, EyeOff, Loader2, LockKeyhole, Mail, RotateCw, UserRound } from 'lucide-react'
import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Link, Navigate, useLocation, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { Button } from '../components/ui/button'
import { Input } from '../components/ui/input'
import { portalAPI, PortalAPIError } from '../portal/api'
import { useUserAuth } from '../providers/AppProviders'

const supportedModes = new Set(['login', 'register', 'check-email', 'verify-email', 'forgot-password', 'reset-password'])

export default function AuthPage() {
  const { mode = 'login' } = useParams()
  const auth = useUserAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [searchParams] = useSearchParams()
  const [email, setEmail] = useState(() => searchParams.get('email') || '')
  const [password, setPassword] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [remember, setRemember] = useState(true)
  const [acceptTerms, setAcceptTerms] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  const target = useMemo(() => {
    const state = location.state as { from?: string; email?: string } | null
    return state?.from?.startsWith('/console') ? state.from : '/console'
  }, [location.state])

  useEffect(() => {
    const state = location.state as { email?: string } | null
    if (state?.email) setEmail((current) => current || state.email || '')
  }, [location.state])

  useEffect(() => {
    const titles: Record<string, string> = { login: '登录', register: '注册', 'check-email': '查收邮件', 'verify-email': '验证邮箱', 'forgot-password': '找回密码', 'reset-password': '重置密码' }
    document.title = `${titles[mode] || '账号'} · AxisRelay`
    setError('')
    setMessage('')
  }, [mode])

  if (!supportedModes.has(mode)) return <Navigate to="/auth/login" replace />
  if (auth.status === 'authenticated' && mode !== 'reset-password') return <Navigate to={target} replace />

  const submitCredentials = async (event: FormEvent) => {
    event.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      if (mode === 'register') {
        const result = await portalAPI.register({ email, password, display_name: displayName, accept_terms: acceptTerms })
        navigate(`/auth/check-email?email=${encodeURIComponent(result.email)}`, { replace: true, state: { email: result.email } })
      } else {
        const session = await portalAPI.login({ email, password, remember })
        auth.setSession(session)
        navigate(target, { replace: true })
      }
    } catch (reason) {
      if (reason instanceof PortalAPIError && reason.code === 'email_unverified') {
        navigate(`/auth/check-email?email=${encodeURIComponent(email)}`, { state: { email } })
        return
      }
      setError(reason instanceof PortalAPIError ? reason.message : '请求失败，请稍后重试')
    } finally {
      setSubmitting(false)
    }
  }

  const submitEmailAction = async (event: FormEvent) => {
    event.preventDefault()
    setError('')
    setMessage('')
    setSubmitting(true)
    try {
      const result = mode === 'forgot-password' ? await portalAPI.forgotPassword(email) : await portalAPI.resendVerification(email)
      setMessage(result.message)
    } catch (reason) {
      setError(reason instanceof PortalAPIError ? reason.message : '请求失败，请稍后重试')
    } finally {
      setSubmitting(false)
    }
  }

  if (mode === 'check-email') return <EmailNotice email={email} submitting={submitting} error={error} message={message} onEmail={setEmail} onSubmit={submitEmailAction} />
  if (mode === 'verify-email') return <VerifyEmail token={searchParams.get('token') || ''} auth={auth} onDone={() => navigate('/console', { replace: true })} />
  if (mode === 'forgot-password') return <SimpleEmailForm email={email} setEmail={setEmail} submitting={submitting} error={error} message={message} onSubmit={submitEmailAction} />
  if (mode === 'reset-password') return <ResetPassword token={searchParams.get('token') || ''} />

  const registerMode = mode === 'register'
  return (
    <div className="w-full max-w-[420px]">
      <div><p className="text-sm font-semibold text-primary">{registerMode ? '开始使用' : '欢迎回来'}</p><h1 className="mt-2 text-3xl font-bold tracking-tight">{registerMode ? '创建 AxisRelay 账号' : '登录用户中心'}</h1><p className="mt-3 text-sm leading-6 text-muted-foreground">{registerMode ? '验证邮箱后，即可创建 API Key、查看用量并管理账号安全。' : '使用注册邮箱与密码继续。'}</p></div>
      <form className="mt-8 space-y-4" onSubmit={submitCredentials}>
        {registerMode && <Field label="昵称" icon={UserRound}><Input value={displayName} onChange={(event) => setDisplayName(event.target.value)} placeholder="你的称呼" autoComplete="name" required minLength={2} maxLength={64} className="h-11 pl-10" /></Field>}
        <Field label="邮箱" icon={Mail}><Input type="email" value={email} onChange={(event) => setEmail(event.target.value)} placeholder="name@example.com" autoComplete="email" required maxLength={320} className="h-11 pl-10" /></Field>
        <Field label="密码" icon={LockKeyhole} suffix={<PasswordToggle shown={showPassword} onToggle={() => setShowPassword((value) => !value)} />}><Input type={showPassword ? 'text' : 'password'} value={password} onChange={(event) => setPassword(event.target.value)} placeholder={registerMode ? '至少 10 位，含大小写字母和数字' : '输入密码'} autoComplete={registerMode ? 'new-password' : 'current-password'} required minLength={10} maxLength={128} className="h-11 px-10" /></Field>
        {registerMode ? <div className="flex items-start gap-3 text-sm text-muted-foreground"><input id="accept-terms" type="checkbox" checked={acceptTerms} onChange={(event) => setAcceptTerms(event.target.checked)} className="mt-0.5 size-4 shrink-0 cursor-pointer accent-primary" /><span><label htmlFor="accept-terms" className="cursor-pointer">我已阅读并同意 </label><Link to="/docs#terms" className="text-primary hover:underline">服务条款</Link> 与隐私说明。</span></div> : <div className="flex items-center justify-between text-sm"><label className="flex cursor-pointer items-center gap-2 text-muted-foreground"><input type="checkbox" checked={remember} onChange={(event) => setRemember(event.target.checked)} className="accent-primary" />保持登录</label><Link to="/auth/forgot-password" className="font-medium text-primary hover:underline">忘记密码？</Link></div>}
        <Feedback error={error} />
        <Button type="submit" className="h-11 w-full" disabled={submitting || (registerMode && !acceptTerms)}>{submitting ? <Loader2 className="animate-spin" /> : <ArrowRight />}{submitting ? '正在处理…' : registerMode ? '创建账号' : '登录'}</Button>
      </form>
      <div className="mt-7 border-t border-border pt-6 text-center text-sm text-muted-foreground">{registerMode ? '已有账号？' : '还没有账号？'} <Link to={registerMode ? '/auth/login' : '/auth/register'} className="font-semibold text-primary hover:underline">{registerMode ? '直接登录' : '免费注册'}</Link></div>
    </div>
  )
}

function EmailNotice({ email, submitting, error, message, onEmail, onSubmit }: { email: string; submitting: boolean; error: string; message: string; onEmail: (value: string) => void; onSubmit: (event: FormEvent) => void }) {
  return <div className="w-full max-w-[440px] text-center"><span className="mx-auto flex size-14 items-center justify-center rounded-2xl bg-primary/10 text-primary"><Mail /></span><h1 className="mt-6 text-3xl font-bold">请查收验证邮件</h1><p className="mt-3 text-sm leading-6 text-muted-foreground">我们已向 <strong className="text-foreground">{email || '你的邮箱'}</strong> 发送验证链接。链接 24 小时内有效。</p><form onSubmit={onSubmit} className="mt-7 space-y-3 text-left"><Field label="需要重新发送？" icon={Mail}><Input type="email" value={email} onChange={(event) => onEmail(event.target.value)} required className="h-11 pl-10" /></Field><Feedback error={error} message={message} /><Button type="submit" variant="outline" className="h-11 w-full" disabled={submitting}>{submitting ? <Loader2 className="animate-spin" /> : <RotateCw />}重新发送验证邮件</Button></form><p className="mt-6 text-sm text-muted-foreground">邮箱写错了？<Link to="/auth/register" className="ml-1 font-medium text-primary hover:underline">重新注册</Link></p></div>
}

function VerifyEmail({ token, auth, onDone }: { token: string; auth: ReturnType<typeof useUserAuth>; onDone: () => void }) {
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState(token ? '' : '验证链接缺少必要信息')
  const verify = async () => { setError(''); setSubmitting(true); try { const session = await portalAPI.verifyEmail(token); auth.setSession(session); onDone() } catch (reason) { setError(reason instanceof PortalAPIError ? reason.message : '验证失败，请稍后重试') } finally { setSubmitting(false) } }
  return <div className="w-full max-w-[420px] text-center"><span className="mx-auto flex size-14 items-center justify-center rounded-2xl bg-primary/10 text-primary"><CheckCircle2 /></span><h1 className="mt-6 text-3xl font-bold">确认邮箱验证</h1><p className="mt-3 text-sm leading-6 text-muted-foreground">点击下方按钮完成验证。为保护账号，验证链接只能使用一次。</p><div className="mt-7"><Feedback error={error} /><Button className="mt-3 h-11 w-full" disabled={!token || submitting} onClick={() => void verify()}>{submitting ? <Loader2 className="animate-spin" /> : <ArrowRight />}确认并进入用户中心</Button></div></div>
}

function SimpleEmailForm({ email, setEmail, submitting, error, message, onSubmit }: { email: string; setEmail: (value: string) => void; submitting: boolean; error: string; message: string; onSubmit: (event: FormEvent) => void }) {
  return <div className="w-full max-w-[420px]"><p className="text-sm font-semibold text-primary">账号恢复</p><h1 className="mt-2 text-3xl font-bold">找回密码</h1><p className="mt-3 text-sm leading-6 text-muted-foreground">输入注册邮箱。如果账号存在，我们会发送一封 30 分钟内有效的重置邮件。</p><form className="mt-8 space-y-4" onSubmit={onSubmit}><Field label="邮箱" icon={Mail}><Input type="email" value={email} onChange={(event) => setEmail(event.target.value)} autoComplete="email" required className="h-11 pl-10" /></Field><Feedback error={error} message={message} /><Button type="submit" className="h-11 w-full" disabled={submitting}>{submitting ? <Loader2 className="animate-spin" /> : <ArrowRight />}发送重置邮件</Button></form><Link to="/auth/login" className="mt-6 block text-center text-sm font-medium text-primary hover:underline">返回登录</Link></div>
}

function ResetPassword({ token }: { token: string }) {
  const [password, setPassword] = useState('')
  const [shown, setShown] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState(token ? '' : '重置链接缺少必要信息')
  const [done, setDone] = useState(false)
  const submit = async (event: FormEvent) => { event.preventDefault(); setError(''); setSubmitting(true); try { await portalAPI.resetPassword(token, password); setDone(true) } catch (reason) { setError(reason instanceof PortalAPIError ? reason.message : '重置失败，请稍后重试') } finally { setSubmitting(false) } }
  if (done) return <div className="w-full max-w-[420px] text-center"><CheckCircle2 className="mx-auto size-12 text-emerald-500" /><h1 className="mt-5 text-3xl font-bold">密码已更新</h1><p className="mt-3 text-sm text-muted-foreground">为保护账号，其他设备上的旧会话已全部退出。</p><Button asChild className="mt-7 h-11 w-full"><Link to="/auth/login">使用新密码登录</Link></Button></div>
  return <div className="w-full max-w-[420px]"><p className="text-sm font-semibold text-primary">账号恢复</p><h1 className="mt-2 text-3xl font-bold">设置新密码</h1><p className="mt-3 text-sm text-muted-foreground">新密码至少 10 位，并同时包含大小写字母和数字。</p><form className="mt-8 space-y-4" onSubmit={submit}><Field label="新密码" icon={LockKeyhole} suffix={<PasswordToggle shown={shown} onToggle={() => setShown((value) => !value)} />}><Input type={shown ? 'text' : 'password'} value={password} onChange={(event) => setPassword(event.target.value)} autoComplete="new-password" required minLength={10} maxLength={128} className="h-11 px-10" /></Field><Feedback error={error} /><Button type="submit" className="h-11 w-full" disabled={!token || submitting}>{submitting ? <Loader2 className="animate-spin" /> : <ArrowRight />}更新密码</Button></form></div>
}

function PasswordToggle({ shown, onToggle }: { shown: boolean; onToggle: () => void }) { return <button type="button" className="text-muted-foreground hover:text-foreground" onClick={onToggle} aria-label={shown ? '隐藏密码' : '显示密码'}>{shown ? <EyeOff className="size-4" /> : <Eye className="size-4" />}</button> }
function Feedback({ error, message }: { error?: string; message?: string }) { return <>{error && <div role="alert" className="rounded-lg border border-red-500/20 bg-red-500/8 px-3.5 py-3 text-sm text-red-600 dark:text-red-400">{error}</div>}{message && <div role="status" className="rounded-lg border border-emerald-500/20 bg-emerald-500/8 px-3.5 py-3 text-sm text-emerald-700 dark:text-emerald-400">{message}</div>}</> }
function Field({ label, icon: Icon, suffix, children }: { label: string; icon: typeof Mail; suffix?: React.ReactNode; children: React.ReactNode }) { return <label className="block"><span className="mb-1.5 block text-sm font-medium">{label}</span><span className="relative block"><Icon className="pointer-events-none absolute left-3 top-1/2 z-10 size-4 -translate-y-1/2 text-muted-foreground" />{children}{suffix && <span className="absolute right-3 top-1/2 z-10 -translate-y-1/2">{suffix}</span>}</span></label> }
