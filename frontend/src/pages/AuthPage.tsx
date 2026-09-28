import { ArrowRight, Eye, EyeOff, Loader2, LockKeyhole, Mail, UserRound } from 'lucide-react'
import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Link, Navigate, useLocation, useNavigate, useParams } from 'react-router-dom'
import { Button } from '../components/ui/button'
import { Input } from '../components/ui/input'
import { portalAPI, PortalAPIError } from '../portal/api'
import { useUserAuth } from '../providers/AppProviders'

export default function AuthPage() {
  const { mode = 'login' } = useParams()
  const registerMode = mode === 'register'
  const auth = useUserAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [remember, setRemember] = useState(true)
  const [acceptTerms, setAcceptTerms] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const target = useMemo(() => {
    const state = location.state as { from?: string } | null
    return state?.from?.startsWith('/console') ? state.from : '/console'
  }, [location.state])

  useEffect(() => {
    document.title = `${registerMode ? '注册' : '登录'} · AxisRelay`
    setError('')
  }, [registerMode])

  if (mode !== 'login' && mode !== 'register') return <Navigate to="/auth/login" replace />
  if (auth.status === 'authenticated') return <Navigate to={target} replace />

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      const session = registerMode
        ? await portalAPI.register({ email, password, display_name: displayName, accept_terms: acceptTerms })
        : await portalAPI.login({ email, password, remember })
      auth.setSession(session)
      navigate(target, { replace: true })
    } catch (reason) {
      setError(reason instanceof PortalAPIError ? reason.message : '请求失败，请稍后重试')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="w-full max-w-[420px]">
      <div><p className="text-sm font-semibold text-primary">{registerMode ? 'Create account' : 'Welcome back'}</p><h1 className="mt-2 text-3xl font-bold tracking-tight">{registerMode ? '创建你的 AxisRelay 账号' : '登录用户中心'}</h1><p className="mt-3 text-sm leading-6 text-muted-foreground">{registerMode ? '注册后即可创建 API Key、查看用量和管理安全设置。' : '使用注册邮箱与密码继续。管理员请前往管理后台登录。'}</p></div>
      <form className="mt-8 space-y-4" onSubmit={submit}>
        {registerMode && <Field label="昵称" icon={UserRound}><Input value={displayName} onChange={(event) => setDisplayName(event.target.value)} placeholder="你的称呼" autoComplete="name" required minLength={2} maxLength={64} className="h-11 pl-10" /></Field>}
        <Field label="邮箱" icon={Mail}><Input type="email" value={email} onChange={(event) => setEmail(event.target.value)} placeholder="name@example.com" autoComplete="email" required maxLength={320} className="h-11 pl-10" /></Field>
        <Field label="密码" icon={LockKeyhole} suffix={<button type="button" className="text-muted-foreground hover:text-foreground" onClick={() => setShowPassword((value) => !value)} aria-label={showPassword ? '隐藏密码' : '显示密码'}>{showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}</button>}>
          <Input type={showPassword ? 'text' : 'password'} value={password} onChange={(event) => setPassword(event.target.value)} placeholder={registerMode ? '至少 10 位，含大小写字母和数字' : '输入密码'} autoComplete={registerMode ? 'new-password' : 'current-password'} required minLength={10} maxLength={128} className="h-11 px-10" />
        </Field>
        {registerMode ? (
          <div className="flex items-start gap-3 text-sm text-muted-foreground"><input id="accept-terms" type="checkbox" checked={acceptTerms} onChange={(event) => setAcceptTerms(event.target.checked)} className="mt-0.5 size-4 shrink-0 cursor-pointer accent-primary" /><span><label htmlFor="accept-terms" className="cursor-pointer">我已阅读并同意 </label><Link to="/docs#terms" className="text-primary hover:underline">服务条款</Link> 与隐私说明。</span></div>
        ) : (
          <div className="flex items-center justify-between text-sm"><label className="flex cursor-pointer items-center gap-2 text-muted-foreground"><input type="checkbox" checked={remember} onChange={(event) => setRemember(event.target.checked)} className="accent-primary" />保持登录</label><span className="text-xs text-muted-foreground">忘记密码请联系管理员</span></div>
        )}
        {error && <div role="alert" className="rounded-lg border border-red-500/20 bg-red-500/8 px-3.5 py-3 text-sm text-red-600 dark:text-red-400">{error}</div>}
        <Button type="submit" className="h-11 w-full" disabled={submitting || (registerMode && !acceptTerms)}>{submitting ? <Loader2 className="animate-spin" /> : <ArrowRight />}{submitting ? '正在处理…' : registerMode ? '创建账号' : '登录'}</Button>
      </form>
      <div className="mt-7 border-t border-border pt-6 text-center text-sm text-muted-foreground">{registerMode ? '已有账号？' : '还没有账号？'} <Link to={registerMode ? '/auth/login' : '/auth/register'} className="font-semibold text-primary hover:underline">{registerMode ? '直接登录' : '免费注册'}</Link></div>
      {!registerMode && <div className="mt-4 text-center"><a href="/admin/" className="text-xs text-muted-foreground hover:text-foreground">管理员入口 →</a></div>}
    </div>
  )
}

function Field({ label, icon: Icon, suffix, children }: { label: string; icon: typeof Mail; suffix?: React.ReactNode; children: React.ReactNode }) {
  return <label className="block"><span className="mb-1.5 block text-sm font-medium">{label}</span><span className="relative block"><Icon className="pointer-events-none absolute left-3 top-1/2 z-10 size-4 -translate-y-1/2 text-muted-foreground" />{children}{suffix && <span className="absolute right-3 top-1/2 z-10 -translate-y-1/2">{suffix}</span>}</span></label>
}
