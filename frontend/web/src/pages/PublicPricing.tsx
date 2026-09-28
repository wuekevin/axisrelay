import { ArrowRight, Check, CircleDollarSign, ShieldCheck } from 'lucide-react'
import { Link } from 'react-router-dom'
import { Button } from '../components/ui/button'

const plans = [
  { name: '免费版', price: '¥0', period: '长期', description: '用于接入验证和个人开发。注册并验证邮箱后即可创建 API Key。', cta: '免费注册', href: '/auth/register', featured: false, features: ['最多 10 个 API Key', '用量与请求明细', '开放的模型协议', '基础使用支持'] },
  { name: '按量版', price: '按量', period: '透明计费', description: '按实际模型用量结算，不收固定席位费。具体单价与可用额度以用户中心展示为准。', cta: '开始使用', href: '/auth/register', featured: true, features: ['按模型精确计费', '实时额度与用量', '流式与工具调用', '自动故障切换'] },
  { name: '团队版', price: '定制', period: '企业方案', description: '适合需要团队额度、专属模型范围和接入支持的组织。', cta: '了解企业支持', href: '/docs#support', featured: false, features: ['团队接入方案', '专属模型范围', '组织级额度控制', '迁移与使用支持'] },
]

export default function PublicPricing() {
  return (
    <div>
      <section className="border-b border-border bg-gradient-to-b from-primary/8 to-background"><div className="mx-auto max-w-4xl px-4 py-20 text-center sm:px-6"><p className="text-xs font-bold uppercase tracking-[.22em] text-primary">清晰价格</p><h1 className="mt-4 text-4xl font-bold tracking-tight sm:text-5xl">从试用到生产，成本始终清晰</h1><p className="mx-auto mt-5 max-w-2xl text-lg leading-8 text-muted-foreground">模型用量、额度和请求明细集中展示，让每一笔消耗都更容易理解。</p></div></section>
      <section className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
        <div className="grid gap-5 lg:grid-cols-3">
          {plans.map((plan) => <article key={plan.name} className={`relative flex flex-col rounded-2xl border p-6 ${plan.featured ? 'border-primary bg-primary/[.045] shadow-xl shadow-primary/10' : 'border-border bg-card shadow-sm'}`}>{plan.featured && <span className="absolute right-5 top-5 rounded-full bg-primary px-2.5 py-1 text-[11px] font-bold text-primary-foreground">推荐</span>}<div><p className="text-sm font-semibold text-primary">{plan.name}</p><div className="mt-4 flex items-end gap-2"><span className="text-4xl font-bold tracking-tight">{plan.price}</span><span className="pb-1 text-sm text-muted-foreground">{plan.period}</span></div><p className="mt-5 min-h-20 text-sm leading-6 text-muted-foreground">{plan.description}</p></div><ul className="my-7 grid flex-1 gap-3">{plan.features.map((item) => <li key={item} className="flex items-center gap-2.5 text-sm"><span className="flex size-5 items-center justify-center rounded-full bg-emerald-500/10"><Check className="size-3.5 text-emerald-500" /></span>{item}</li>)}</ul><Button variant={plan.featured ? 'default' : 'outline'} className="h-11" asChild><Link to={plan.href}>{plan.cta}<ArrowRight /></Link></Button></article>)}
        </div>
        <div className="mt-10 grid gap-4 md:grid-cols-3">
          <Info icon={CircleDollarSign} title="美元计费口径" text="用户中心按网关结算口径展示实际消耗，避免模型单位和汇率混淆。" />
          <Info icon={ShieldCheck} title="额度控制" text="达到已结算的 Key 配额后拒绝新请求；并发与异步请求仍应预留安全余量。" />
          <Info icon={Check} title="随时撤销" text="API Key 可以随时撤销，历史用量仍保留在审计记录中。" />
        </div>
      </section>
      <section className="border-t border-border bg-card/40"><div className="mx-auto max-w-4xl px-4 py-16 sm:px-6"><h2 className="text-center text-3xl font-bold">常见问题</h2><div className="mt-10 grid gap-4">{[
        ['注册后有多少试用额度？', '用户中心创建 Key 时会显示当前可用额度和该 Key 的额度上限。'],
        ['模型价格会变化吗？', '模型价格可能随服务调整；当前用量和结算口径会在用户中心清晰展示。'],
        ['可以限制单个 Key 吗？', '可以。网关支持总额度、请求速率、并发、模型范围和时间窗口限制。'],
        ['是否支持团队接入？', '支持。团队可以获得更适合组织使用的模型范围、额度控制与迁移支持。'],
      ].map(([q, a]) => <details key={q} className="group rounded-xl border border-border bg-background p-5"><summary className="cursor-pointer list-none font-semibold">{q}</summary><p className="mt-3 text-sm leading-6 text-muted-foreground">{a}</p></details>)}</div></div></section>
    </div>
  )
}

function Info({ icon: Icon, title, text }: { icon: typeof Check; title: string; text: string }) {
  return <div className="rounded-xl border border-border bg-card p-5"><Icon className="size-5 text-primary" /><h3 className="mt-3 font-semibold">{title}</h3><p className="mt-2 text-sm leading-6 text-muted-foreground">{text}</p></div>
}
