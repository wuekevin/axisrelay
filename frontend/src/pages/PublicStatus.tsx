import { Activity, CheckCircle2, Clock3, Database, RefreshCw, Server, Wifi } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Button } from '../components/ui/button'

type Health = { status: string; build_version?: string; available?: number; total?: number; counts_complete?: boolean }

export default function PublicStatus() {
  const [health, setHealth] = useState<Health | null>(null)
  const [loading, setLoading] = useState(true)
  const [checkedAt, setCheckedAt] = useState<Date | null>(null)
  const load = async () => {
    setLoading(true)
    try {
      const response = await fetch('/health', { cache: 'no-store' })
      const payload = await response.json()
      setHealth(response.ok ? payload : { status: 'unavailable' })
    } catch { setHealth({ status: 'unavailable' }) }
    finally { setCheckedAt(new Date()); setLoading(false) }
  }
  useEffect(() => { void load() }, [])
  const healthy = health?.status === 'ok'
  return <div><section className="border-b border-border bg-gradient-to-b from-emerald-500/8 to-background"><div className="mx-auto max-w-4xl px-4 py-20 text-center sm:px-6"><Activity className="mx-auto size-9 text-emerald-500" /><h1 className="mt-5 text-4xl font-bold tracking-tight">服务状态</h1><p className="mx-auto mt-4 max-w-2xl text-lg text-muted-foreground">查看当前网关进程与账号池的实时健康状态。</p></div></section><section className="mx-auto max-w-4xl px-4 py-14 sm:px-6"><div className={`rounded-2xl border p-6 ${healthy ? 'border-emerald-500/20 bg-emerald-500/[.055]' : 'border-red-500/20 bg-red-500/[.055]'}`}><div className="flex flex-col justify-between gap-5 sm:flex-row sm:items-center"><div className="flex items-center gap-4">{healthy ? <CheckCircle2 className="size-9 text-emerald-500" /> : <Activity className="size-9 text-red-500" />}<div><h2 className="text-xl font-bold">{loading ? '正在检查…' : healthy ? '所有核心服务运行正常' : '服务当前不可用'}</h2><p className="mt-1 text-sm text-muted-foreground">{checkedAt ? `最后检查：${checkedAt.toLocaleString()}` : '正在连接健康检查端点'}</p></div></div><Button variant="outline" onClick={() => void load()} disabled={loading}><RefreshCw className={loading ? 'animate-spin' : ''} />重新检查</Button></div></div><div className="mt-6 grid gap-4 sm:grid-cols-2"><StatusCard icon={Server} title="API 网关" value={healthy ? '正常' : '异常'} healthy={healthy} /><StatusCard icon={Wifi} title="可用上游账号" value={health?.counts_complete ? `${health.available ?? 0} / ${health.total ?? 0}` : '统计中'} healthy={healthy} /><StatusCard icon={Database} title="持久化服务" value={healthy ? '已连接' : '未知'} healthy={healthy} /><StatusCard icon={Clock3} title="构建版本" value={health?.build_version || '未报告'} healthy={healthy} /></div><div className="mt-10 rounded-xl border border-border bg-card p-5"><h3 className="font-semibold">状态说明</h3><p className="mt-2 text-sm leading-6 text-muted-foreground">此页面直接读取当前实例的 `/health` 结果。它证明网关进程和核心账号池锁可响应，但不代表所有第三方模型提供商在所有地区都可用。具体请求状态请在用户中心查看。</p></div></section></div>
}

function StatusCard({ icon: Icon, title, value, healthy }: { icon: typeof Server; title: string; value: string; healthy: boolean }) { return <div className="flex items-center gap-4 rounded-xl border border-border bg-card p-5"><span className="flex size-10 items-center justify-center rounded-lg bg-muted"><Icon className="size-5 text-primary" /></span><div><div className="text-xs text-muted-foreground">{title}</div><div className={`mt-1 font-semibold ${healthy ? '' : 'text-muted-foreground'}`}>{value}</div></div></div> }
