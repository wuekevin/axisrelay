import { ArrowRight, Check, CircleDollarSign, ShieldCheck } from 'lucide-react'
import { Link } from 'react-router-dom'
import { Button } from '../components/ui/button'

const plans = [
  { name: 'Free', price: '¥0', period: '长期', description: '用于接入验证和个人开发。注册后可创建 API Key，并获得由部署方配置的试用额度。', cta: '免费注册', href: '/auth/register', featured: false, features: ['最多 10 个 API Key', '用量与请求明细', '全部开放模型协议', '社区级支持'] },
  { name: 'Usage', price: '按量', period: '透明计费', description: '按真实模型用量结算，不收固定席位费。具体单价以用户中心和管理员发布的价格为准。', cta: '开始使用', href: '/auth/register', featured: true, features: ['按模型精确计费', '实时余额与用量', '流式与工具调用', '自动故障切换'] },
  { name: 'Enterprise', price: '定制', period: '企业方案', description: '适合需要私有部署、专属账号池、配额策略和运维支持的团队。', cta: '联系管理员', href: '/docs#support', featured: false, features: ['私有化部署', '专属模型与账号池', '组织级配额策略', '迁移与运维支持'] },
]

export default function PublicPricing() {
  return (
    <div>
      <section className="border-b border-border bg-gradient-to-b from-primary/8 to-background"><div className="mx-auto max-w-4xl px-4 py-20 text-center sm:px-6"><p className="text-xs font-bold uppercase tracking-[.22em] text-primary">Simple pricing</p><h1 className="mt-4 text-4xl font-bold tracking-tight sm:text-5xl">从试用到生产，成本始终清晰</h1><p className="mx-auto mt-5 max-w-2xl text-lg leading-8 text-muted-foreground">不隐藏模型用量，不把平台费混进 Token。管理员可为不同部署设置额度和结算策略。</p></div></section>
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
        ['注册后有多少试用额度？', '默认试用额度由部署管理员配置；用户中心创建 Key 时会显示该 Key 的额度上限。'],
        ['模型价格会变化吗？', '上游价格可能调整。AxisRelay 会以请求发生时的价格快照记录用量，部署方应在价格变化时同步公告。'],
        ['可以限制单个 Key 吗？', '可以。网关支持总额度、请求速率、并发、模型范围和时间窗口限制。'],
        ['是否支持私有部署？', '支持。生产包使用 MySQL、Redis 和本地构建镜像，适合放在自有网络与反向代理之后。'],
      ].map(([q, a]) => <details key={q} className="group rounded-xl border border-border bg-background p-5"><summary className="cursor-pointer list-none font-semibold">{q}</summary><p className="mt-3 text-sm leading-6 text-muted-foreground">{a}</p></details>)}</div></div></section>
    </div>
  )
}

function Info({ icon: Icon, title, text }: { icon: typeof Check; title: string; text: string }) {
  return <div className="rounded-xl border border-border bg-card p-5"><Icon className="size-5 text-primary" /><h3 className="mt-3 font-semibold">{title}</h3><p className="mt-2 text-sm leading-6 text-muted-foreground">{text}</p></div>
}
