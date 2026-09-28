import {
  Activity, ArrowRight, Braces, Check, ChevronRight, CircleGauge, CloudCog,
  Code2, DatabaseZap, Gauge, KeyRound, Layers3, LockKeyhole, Network,
  Route, ShieldCheck, Sparkles, TerminalSquare, Waypoints,
} from 'lucide-react'
import { Link } from 'react-router-dom'
import { Button } from '../components/ui/button'
import { useUserAuth } from '../providers/AppProviders'

const providers = [
  ['OpenAI', 'Responses / Chat Completions'],
  ['Anthropic', 'Messages / Claude Code'],
  ['Google', 'Gemini native streaming'],
  ['xAI', 'Grok text and tools'],
]

const features = [
  { icon: Route, title: '统一协议入口', text: 'OpenAI、Anthropic 与 Gemini 兼容端点共用一套鉴权和观测体系。' },
  { icon: Waypoints, title: '智能账号调度', text: '并发感知、会话亲和、冷却与故障转移，让请求自动选择可用上游。' },
  { icon: ShieldCheck, title: '生产级安全', text: '密钥隔离、凭据加密、请求限制与安全审计默认进入运行链路。' },
  { icon: CircleGauge, title: '实时用量可见', text: '按 Key 查看请求、Token、模型、延迟和错误，不再依赖模糊估算。' },
  { icon: DatabaseZap, title: '可靠数据层', text: 'MySQL 持久化关键配置与账单数据，Redis 承担低延迟缓存。' },
  { icon: CloudCog, title: '面向运维设计', text: '健康检查、结构化错误、恢复策略与可回滚部署配置一并提供。' },
]

export default function PublicHome() {
  const auth = useUserAuth()
  return (
    <>
      <section className="relative overflow-hidden border-b border-border/60">
        <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_20%_10%,color-mix(in_srgb,var(--color-primary)_18%,transparent),transparent_32%),radial-gradient(circle_at_80%_5%,color-mix(in_srgb,var(--color-primary)_10%,transparent),transparent_26%)]" />
        <div className="relative mx-auto grid max-w-7xl items-center gap-14 px-4 py-20 sm:px-6 sm:py-28 lg:grid-cols-[1.08fr_.92fr] lg:px-8 lg:py-32">
          <div>
            <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/8 px-3 py-1.5 text-xs font-semibold text-primary">
              <Sparkles className="size-3.5" /> 为多模型生产流量而生
            </div>
            <h1 className="max-w-4xl text-4xl font-bold tracking-[-0.045em] sm:text-6xl lg:text-7xl">
              一套 API，连接你的
              <span className="block bg-gradient-to-r from-primary via-cyan-500 to-emerald-500 bg-clip-text text-transparent">全部 AI 能力</span>
            </h1>
            <p className="mt-7 max-w-2xl text-lg leading-8 text-muted-foreground sm:text-xl">AxisRelay 把模型接入、账号调度、协议转换、用量统计和故障恢复放进同一个可靠网关。你只需要维护一个端点。</p>
            <div className="mt-9 flex flex-col gap-3 sm:flex-row">
              <Button size="lg" className="h-12 px-6 text-base" asChild>
                <Link to={auth.authenticated ? '/console' : '/auth/register'}>{auth.authenticated ? '打开用户中心' : '免费开始使用'} <ArrowRight /></Link>
              </Button>
              <Button size="lg" variant="outline" className="h-12 px-6 text-base" asChild>
                <Link to="/docs"><TerminalSquare /> 查看快速开始</Link>
              </Button>
            </div>
            <div className="mt-8 flex flex-wrap gap-x-6 gap-y-2 text-sm text-muted-foreground">
              {['无需绑定信用卡', '兼容主流 SDK', 'Key 可随时撤销'].map((item) => <span key={item} className="flex items-center gap-1.5"><Check className="size-4 text-emerald-500" />{item}</span>)}
            </div>
          </div>

          <div className="relative mx-auto w-full max-w-xl">
            <div className="absolute -inset-8 rounded-[40px] bg-primary/10 blur-3xl" />
            <div className="relative overflow-hidden rounded-2xl border border-border bg-[#0b1020] shadow-2xl shadow-primary/10">
              <div className="flex h-11 items-center gap-2 border-b border-white/10 px-4">
                <span className="size-2.5 rounded-full bg-red-400" /><span className="size-2.5 rounded-full bg-amber-400" /><span className="size-2.5 rounded-full bg-emerald-400" />
                <span className="ml-2 text-xs text-slate-500">first-request.ts</span>
              </div>
              <pre className="overflow-x-auto p-5 text-[13px] leading-6 text-slate-300 sm:p-7 sm:text-sm"><code><span className="text-fuchsia-300">import</span> OpenAI <span className="text-fuchsia-300">from</span> <span className="text-emerald-300">'openai'</span>{'\n\n'}<span className="text-fuchsia-300">const</span> client = <span className="text-fuchsia-300">new</span> OpenAI({'{'}{'\n'}  baseURL: <span className="text-emerald-300">'https://your-domain.com/v1'</span>,{'\n'}  apiKey: process.env.<span className="text-cyan-300">AXISRELAY_API_KEY</span>{'\n'}{'}'}){'\n\n'}<span className="text-fuchsia-300">const</span> response = <span className="text-fuchsia-300">await</span> client.responses.create({'{'}{'\n'}  model: <span className="text-emerald-300">'gpt-5.4'</span>,{'\n'}  input: <span className="text-emerald-300">'Design a resilient API.'</span>{'\n'}{'}'})</code></pre>
              <div className="grid grid-cols-3 border-t border-white/10 bg-white/[.025]">
                {[['SSE / WS', '流式协议'], ['实时', '用量观测'], ['4', '上游渠道']].map(([value, label]) => <div key={label} className="border-r border-white/10 p-4 last:border-0"><div className="font-mono text-sm font-semibold text-white">{value}</div><div className="mt-1 text-[11px] text-slate-500">{label}</div></div>)}
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="border-b border-border/60 bg-card/30">
        <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
          <p className="text-center text-xs font-semibold uppercase tracking-[.2em] text-muted-foreground">一个网关覆盖主流模型生态</p>
          <div className="mt-7 grid gap-px overflow-hidden rounded-xl border border-border bg-border sm:grid-cols-2 lg:grid-cols-4">
            {providers.map(([name, caption]) => <div key={name} className="bg-card px-5 py-4"><div className="text-sm font-semibold">{name}</div><div className="mt-1 text-xs text-muted-foreground">{caption}</div></div>)}
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-4 py-20 sm:px-6 sm:py-28 lg:px-8">
        <SectionHeading eyebrow="Platform" title="从第一行代码到生产流量" description="需要稳定运行的部分已经被整合进网关，你的团队可以专注于业务体验。" />
        <div className="mt-12 grid gap-5 md:grid-cols-2 lg:grid-cols-3">
          {features.map(({ icon: Icon, title, text }) => (
            <div key={title} className="group rounded-2xl border border-border bg-card p-6 shadow-sm transition hover:-translate-y-1 hover:border-primary/30 hover:shadow-lg">
              <div className="flex size-11 items-center justify-center rounded-xl bg-primary/10 text-primary"><Icon className="size-5" /></div>
              <h3 className="mt-5 text-lg font-semibold">{title}</h3>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">{text}</p>
            </div>
          ))}
        </div>
      </section>

      <section className="border-y border-border bg-card/40">
        <div className="mx-auto grid max-w-7xl gap-14 px-4 py-20 sm:px-6 sm:py-28 lg:grid-cols-2 lg:px-8">
          <div>
            <SectionHeading eyebrow="One endpoint" title="协议可以不同，接入不必复杂" description="使用熟悉的 SDK 和请求格式。AxisRelay 在内部完成模型映射、协议转换与流式事件处理。" align="left" />
            <div className="mt-8 grid gap-4">
              {[
                [Braces, 'OpenAI Responses 与 Chat Completions'],
                [Code2, 'Anthropic Messages 协议'],
                [Layers3, 'Gemini 原生内容与工具调用'],
                [Network, 'HTTP SSE 与 WebSocket 流式传输'],
              ].map(([Icon, label]) => {
                const IconComponent = Icon as typeof Braces
                return <div key={label as string} className="flex items-center gap-3 rounded-xl border border-border bg-background px-4 py-3"><IconComponent className="size-4 text-primary" /><span className="text-sm font-medium">{label as string}</span></div>
              })}
            </div>
          </div>
          <div className="rounded-2xl border border-border bg-background p-3 shadow-xl">
            <div className="rounded-xl border border-border bg-card p-6">
              <div className="flex items-center justify-between"><span className="text-sm font-semibold">请求调度链路</span><span className="rounded-full bg-emerald-500/10 px-2.5 py-1 text-xs font-medium text-emerald-600">Healthy</span></div>
              <div className="mt-8 space-y-3">
                {[
                  [KeyRound, 'API Key 鉴权与配额', '通过'],
                  [LockKeyhole, '安全策略与请求校验', '通过'],
                  [Route, '模型路由与账号选择', '动态'],
                  [Activity, '流式响应与用量结算', '实时'],
                ].map(([Icon, title, value], index) => {
                  const IconComponent = Icon as typeof KeyRound
                  return <div key={title as string} className="relative flex items-center gap-4 rounded-xl bg-muted/60 p-4"><div className="flex size-9 items-center justify-center rounded-lg bg-card text-primary shadow-sm"><IconComponent className="size-4" /></div><div className="min-w-0 flex-1"><div className="text-sm font-medium">{title as string}</div><div className="text-xs text-muted-foreground">步骤 {index + 1}</div></div><span className="font-mono text-xs text-muted-foreground">{value as string}</span></div>
                })}
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-5xl px-4 py-20 text-center sm:px-6 sm:py-28">
        <div className="rounded-3xl border border-primary/20 bg-gradient-to-br from-primary/12 via-card to-cyan-500/8 px-6 py-14 shadow-xl shadow-primary/5 sm:px-12">
          <Gauge className="mx-auto size-9 text-primary" />
          <h2 className="mt-5 text-3xl font-bold tracking-tight sm:text-4xl">准备好简化你的 AI 基础设施了吗？</h2>
          <p className="mx-auto mt-4 max-w-2xl text-muted-foreground">注册账号，创建第一把 API Key，在几分钟内完成现有 SDK 的端点切换。</p>
          <Button size="lg" className="mt-8 h-12 px-7" asChild><Link to={auth.authenticated ? '/console/keys' : '/auth/register'}>{auth.authenticated ? '管理 API Key' : '创建免费账号'} <ChevronRight /></Link></Button>
        </div>
      </section>
    </>
  )
}

function SectionHeading({ eyebrow, title, description, align = 'center' }: { eyebrow: string; title: string; description: string; align?: 'left' | 'center' }) {
  return <div className={align === 'center' ? 'mx-auto max-w-3xl text-center' : 'max-w-2xl'}><p className="text-xs font-bold uppercase tracking-[.22em] text-primary">{eyebrow}</p><h2 className="mt-3 text-3xl font-bold tracking-tight sm:text-4xl">{title}</h2><p className="mt-4 text-base leading-7 text-muted-foreground sm:text-lg">{description}</p></div>
}
