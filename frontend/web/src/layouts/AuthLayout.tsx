import { ArrowLeft, CheckCircle2, Network, ShieldCheck, Sparkles } from 'lucide-react'
import { Link, Outlet } from 'react-router-dom'
import PortalLogo from '../portal/PortalLogo'

export default function AuthLayout() {
  return (
    <div className="grid min-h-dvh bg-background lg:grid-cols-[1.05fr_.95fr]">
      <section className="relative hidden overflow-hidden border-r border-border bg-[#09101f] text-white lg:flex lg:flex-col">
        <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_20%_15%,rgba(14,165,233,.22),transparent_34%),radial-gradient(circle_at_80%_85%,rgba(16,185,129,.15),transparent_32%)]" />
        <div className="relative z-10 flex h-20 items-center px-10"><PortalLogo className="text-white" /></div>
        <div className="relative z-10 flex flex-1 flex-col justify-center px-10 pb-20 xl:px-20">
          <span className="mb-6 flex size-12 items-center justify-center rounded-2xl border border-white/10 bg-white/10"><Sparkles className="size-5 text-cyan-300" /></span>
          <h1 className="max-w-xl text-4xl font-bold leading-tight tracking-tight xl:text-5xl">让模型接入保持简单，让生产运行足够可靠。</h1>
          <p className="mt-5 max-w-xl text-base leading-7 text-slate-400">创建一个账号，即可管理真实可用的 API Key、查看模型用量并控制所有登录设备。</p>
          <div className="mt-10 grid max-w-xl gap-4 sm:grid-cols-3">
            {[[Network, '统一端点'], [ShieldCheck, '安全会话'], [CheckCircle2, '透明用量']].map(([Icon, label]) => { const Component = Icon as typeof Network; return <div key={label as string} className="rounded-xl border border-white/10 bg-white/[.045] p-4"><Component className="size-4 text-cyan-300" /><div className="mt-3 text-sm font-medium">{label as string}</div></div> })}
          </div>
        </div>
        <div className="relative z-10 px-10 pb-8 text-xs text-slate-500">安全访问 · 清晰用量 · 随时掌控</div>
      </section>
      <section className="flex min-h-dvh flex-col">
        <div className="flex h-20 items-center justify-between px-5 sm:px-8"><div className="lg:hidden"><PortalLogo /></div><Link to="/" className="ml-auto inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"><ArrowLeft className="size-4" /> 返回官网</Link></div>
        <div className="flex flex-1 items-center justify-center px-5 pb-16 sm:px-8"><Outlet /></div>
      </section>
    </div>
  )
}
