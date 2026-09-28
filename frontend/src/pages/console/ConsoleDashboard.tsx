import { Activity, ArrowRight, BarChart3, CheckCircle2, Clock3, KeyRound, Sparkles, WalletCards } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../components/ui/button'
import { portalAPI } from '../../portal/api'
import { formatNumber, formatUSD } from '../../portal/format'
import type { PortalAPIKey, PortalSubscription, PortalUsageReport } from '../../portal/types'
import { useUserAuth } from '../../providers/AppProviders'

export default function ConsoleDashboard() {
  const auth = useUserAuth()
  const [keys, setKeys] = useState<PortalAPIKey[]>([])
  const [subscription, setSubscription] = useState<PortalSubscription | null>(null)
  const [usage, setUsage] = useState<PortalUsageReport | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let active = true
    void Promise.all([portalAPI.apiKeys(), portalAPI.subscription()]).then(async ([keyResponse, subscriptionResponse]) => {
      if (!active) return
      setKeys(keyResponse.api_keys)
      setSubscription(subscriptionResponse.subscription)
      if (keyResponse.api_keys[0]) {
        const report = await portalAPI.usage(keyResponse.api_keys[0].id, 30).catch(() => null)
        if (active && report) setUsage(report.usage)
      }
    }).catch((reason) => {
      if (active) setError(reason instanceof Error ? reason.message : '用户中心数据读取失败')
    }).finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [])

  const quotaLimit = keys.reduce((sum, item) => sum + item.quota_limit, 0)
  const quotaUsed = keys.reduce((sum, item) => sum + item.quota_used, 0)
  return (
    <div className="space-y-7">
      <section className="relative overflow-hidden rounded-2xl border border-border bg-card p-6 shadow-sm sm:p-8">
        <div className="pointer-events-none absolute right-0 top-0 size-64 translate-x-1/3 -translate-y-1/3 rounded-full bg-primary/10 blur-3xl" />
        <div className="relative flex flex-col justify-between gap-6 sm:flex-row sm:items-center"><div><div className="flex items-center gap-2 text-sm font-semibold text-primary"><Sparkles className="size-4" />欢迎回来</div><h2 className="mt-2 text-2xl font-bold sm:text-3xl">{auth.user?.display_name || auth.user?.email}</h2><p className="mt-2 text-sm text-muted-foreground">你的网关凭据、用量和安全状态都集中在这里。</p></div><Button asChild><Link to="/console/keys">管理 API Key <ArrowRight /></Link></Button></div>
      </section>
      {error && <div role="alert" className="rounded-xl border border-red-500/20 bg-red-500/8 p-4 text-sm text-red-600 dark:text-red-400">{error}</div>}

      <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <Metric icon={KeyRound} label="有效 API Key" value={loading ? '—' : String(keys.length)} caption={`最多 ${10} 个`} />
        <Metric icon={BarChart3} label="近 30 天请求" value={loading ? '—' : formatNumber(usage?.summary.requests)} caption={`${formatNumber(usage?.summary.tokens)} tokens`} />
        <Metric icon={WalletCards} label="已用额度" value={loading ? '—' : formatUSD(quotaUsed)} caption={quotaLimit > 0 ? `上限 ${formatUSD(quotaLimit)}` : '暂无额度'} />
        <Metric icon={Activity} label="请求成功率" value={loading || !usage?.summary.requests ? '—' : `${Math.max(0, 100 - usage.summary.error_count / usage.summary.requests * 100).toFixed(1)}%`} caption={`${formatNumber(usage?.summary.error_count)} 次错误`} />
      </section>

      <section className="grid gap-5 lg:grid-cols-[1.35fr_.65fr]">
        <div className="rounded-2xl border border-border bg-card p-6 shadow-sm"><div className="flex items-center justify-between"><div><h3 className="font-semibold">快速开始</h3><p className="mt-1 text-sm text-muted-foreground">完成以下步骤即可发出生产请求。</p></div><Clock3 className="size-5 text-muted-foreground" /></div><div className="mt-6 space-y-3"><Step done={keys.length > 0} number="01" title="创建 API Key" text={keys.length ? `已有 ${keys.length} 个可用 Key` : 'Key 只在创建时完整显示'} href="/console/keys" /><Step done={Boolean(usage?.summary.requests)} number="02" title="发送第一个 API 请求" text="使用 OpenAI 或 Anthropic SDK" href="/docs#first-request" /><Step done={Boolean(usage?.summary.requests)} number="03" title="检查真实用量" text="查看 Token、延迟和错误明细" href="/console/usage" /></div></div>
        <div className="rounded-2xl border border-border bg-card p-6 shadow-sm"><h3 className="font-semibold">当前方案</h3><div className="mt-5 flex size-11 items-center justify-center rounded-xl bg-primary/10 text-primary"><WalletCards /></div><div className="mt-4 text-2xl font-bold">{subscription?.name || 'Free'}</div><div className="mt-1 text-sm capitalize text-emerald-600">{subscription?.status || 'active'}</div><p className="mt-4 text-sm leading-6 text-muted-foreground">API Key 的实际额度在创建时由当前部署策略下发，可在 Key 列表随时查看。</p><Button variant="outline" className="mt-5 w-full" asChild><Link to="/pricing">查看价格说明</Link></Button></div>
      </section>
    </div>
  )
}

function Metric({ icon: Icon, label, value, caption }: { icon: typeof KeyRound; label: string; value: string; caption: string }) { return <div className="rounded-2xl border border-border bg-card p-5 shadow-sm"><div className="flex items-center justify-between"><span className="text-sm text-muted-foreground">{label}</span><Icon className="size-4 text-primary" /></div><div className="mt-3 text-2xl font-bold">{value}</div><div className="mt-1 text-xs text-muted-foreground">{caption}</div></div> }

function Step({ done, number, title, text, href }: { done: boolean; number: string; title: string; text: string; href: string }) { return <Link to={href} className="group flex items-center gap-4 rounded-xl border border-border bg-background p-4 hover:border-primary/30"><span className={`flex size-9 items-center justify-center rounded-lg text-xs font-bold ${done ? 'bg-emerald-500/10 text-emerald-600' : 'bg-muted text-muted-foreground'}`}>{done ? <CheckCircle2 className="size-4" /> : number}</span><span className="min-w-0 flex-1"><span className="block text-sm font-semibold">{title}</span><span className="block truncate text-xs text-muted-foreground">{text}</span></span><ArrowRight className="size-4 text-muted-foreground transition group-hover:translate-x-1 group-hover:text-primary" /></Link> }
