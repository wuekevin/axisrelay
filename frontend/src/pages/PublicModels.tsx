import { Bot, BrainCircuit, Check, Eye, Image, MessageSquareText, Sparkles, Wrench } from 'lucide-react'
import { Link } from 'react-router-dom'
import { Button } from '../components/ui/button'

const families = [
  { provider: 'OpenAI', name: 'GPT & Codex', tone: 'from-emerald-500/20 to-cyan-500/5', description: '适合复杂推理、代码生成、工具调用和多轮代理任务。', capabilities: ['Responses API', '函数调用', '图像理解', '流式输出'] },
  { provider: 'Anthropic', name: 'Claude', tone: 'from-orange-500/20 to-amber-500/5', description: '长上下文、精细写作、代码代理与可靠指令遵循。', capabilities: ['Messages API', 'Extended thinking', '工具调用', 'Prompt cache'] },
  { provider: 'Google', name: 'Gemini', tone: 'from-blue-500/20 to-violet-500/5', description: '原生多模态、超长上下文与 Gemini SDK 兼容输出。', capabilities: ['Gemini 原生协议', '多模态', 'Thought', 'Usage 保留'] },
  { provider: 'xAI', name: 'Grok', tone: 'from-slate-500/20 to-zinc-500/5', description: '文本推理、联网能力与工具密集型实时应用。', capabilities: ['Chat Completions', '工具调用', '连续重试', '模型路由'] },
]

const rows = [
  ['文本与推理', true, true, true, true],
  ['流式响应', true, true, true, true],
  ['函数 / 工具调用', true, true, true, true],
  ['图像理解', true, true, true, false],
  ['图像生成', true, false, true, true],
  ['原生 WebSocket', true, false, false, false],
]

export default function PublicModels() {
  return (
    <div>
      <PageHero eyebrow="Model catalog" title="在一个端点上选择最合适的模型" description="保留各家模型的原生优势，同时用统一的鉴权、限额和观测方式接入。实际可用模型由部署实例的账号池与管理员策略决定。" />
      <section className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
        <div className="grid gap-5 md:grid-cols-2">
          {families.map((family) => (
            <article key={family.name} className="relative overflow-hidden rounded-2xl border border-border bg-card p-6 shadow-sm">
              <div className={`pointer-events-none absolute inset-x-0 top-0 h-32 bg-gradient-to-br ${family.tone}`} />
              <div className="relative">
                <div className="flex items-start justify-between gap-4"><div><p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{family.provider}</p><h2 className="mt-1 text-2xl font-bold">{family.name}</h2></div><div className="flex size-11 items-center justify-center rounded-xl border border-border bg-background/80 text-primary shadow-sm"><Bot /></div></div>
                <p className="mt-5 max-w-xl text-sm leading-6 text-muted-foreground">{family.description}</p>
                <div className="mt-6 flex flex-wrap gap-2">{family.capabilities.map((item) => <span key={item} className="rounded-full border border-border bg-background/70 px-3 py-1.5 text-xs font-medium"><Check className="mr-1 inline size-3 text-emerald-500" />{item}</span>)}</div>
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="border-y border-border bg-card/40">
        <div className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
          <div className="max-w-2xl"><p className="text-xs font-bold uppercase tracking-[.2em] text-primary">Capabilities</p><h2 className="mt-3 text-3xl font-bold tracking-tight">能力矩阵</h2><p className="mt-3 text-muted-foreground">统一入口不会抹平模型差异；AxisRelay 会保留原生终态、工具、思考与用量语义。</p></div>
          <div className="mt-9 overflow-x-auto rounded-2xl border border-border bg-background">
            <table className="w-full min-w-[680px] text-left text-sm">
              <thead className="border-b border-border bg-muted/60"><tr><th className="px-5 py-4 font-semibold">能力</th>{['OpenAI', 'Claude', 'Gemini', 'Grok'].map((name) => <th key={name} className="px-5 py-4 text-center font-semibold">{name}</th>)}</tr></thead>
              <tbody>{rows.map(([name, ...values]) => <tr key={name as string} className="border-b border-border/70 last:border-0"><td className="px-5 py-4 font-medium">{name as string}</td>{values.map((enabled, index) => <td key={index} className="px-5 py-4 text-center">{enabled ? <Check className="mx-auto size-4 text-emerald-500" /> : <span className="text-muted-foreground">—</span>}</td>)}</tr>)}</tbody>
            </table>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {[[BrainCircuit, '推理', '按任务复杂度选择 reasoning effort'], [Wrench, '工具', '保持函数名称与参数结构'], [Eye, '视觉', '统一处理图像输入与引用'], [Image, '生图', '支持同步与异步图片任务']].map(([Icon, title, text]) => { const Component = Icon as typeof BrainCircuit; return <div key={title as string} className="rounded-xl border border-border bg-card p-5"><Component className="size-5 text-primary" /><h3 className="mt-4 font-semibold">{title as string}</h3><p className="mt-2 text-sm leading-6 text-muted-foreground">{text as string}</p></div> })}
        </div>
        <div className="mt-12 flex flex-col items-center rounded-2xl border border-border bg-muted/40 px-6 py-10 text-center"><MessageSquareText className="size-8 text-primary" /><h2 className="mt-4 text-2xl font-bold">不确定该选哪个模型？</h2><p className="mt-2 max-w-xl text-sm text-muted-foreground">从文档中的协议示例开始，或创建 Key 后在用户中心查看真实请求表现。</p><div className="mt-6 flex gap-3"><Button asChild><Link to="/auth/register"><Sparkles /> 免费开始</Link></Button><Button variant="outline" asChild><Link to="/docs">查看文档</Link></Button></div></div>
      </section>
    </div>
  )
}

function PageHero({ eyebrow, title, description }: { eyebrow: string; title: string; description: string }) {
  return <section className="border-b border-border bg-gradient-to-b from-primary/8 to-background"><div className="mx-auto max-w-4xl px-4 py-20 text-center sm:px-6"><p className="text-xs font-bold uppercase tracking-[.22em] text-primary">{eyebrow}</p><h1 className="mt-4 text-4xl font-bold tracking-tight sm:text-5xl">{title}</h1><p className="mx-auto mt-5 max-w-3xl text-lg leading-8 text-muted-foreground">{description}</p></div></section>
}
