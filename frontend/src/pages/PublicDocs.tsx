import { BookOpen, Check, Copy, KeyRound, MessageSquareCode, Rocket, Terminal } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { Button } from '../components/ui/button'

const curlExample = `curl https://your-domain.com/v1/responses \\
  -H "Authorization: Bearer $AXISRELAY_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-5.4",
    "input": "Explain reliable streaming in three steps."
  }'`

const jsExample = `import OpenAI from "openai";

const client = new OpenAI({
  apiKey: process.env.AXISRELAY_API_KEY,
  baseURL: "https://your-domain.com/v1",
});

const response = await client.responses.create({
  model: "gpt-5.4",
  input: "Build a release checklist.",
});

console.log(response.output_text);`

export default function PublicDocs() {
  const { hash } = useLocation()
  useEffect(() => {
    if (!hash) return
    const id = decodeURIComponent(hash.slice(1))
    const frame = window.requestAnimationFrame(() => {
      document.getElementById(id)?.scrollIntoView({ block: 'start' })
    })
    return () => window.cancelAnimationFrame(frame)
  }, [hash])

  return (
    <div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div className="grid gap-10 lg:grid-cols-[240px_minmax(0,1fr)]">
        <aside className="hidden lg:block"><div className="sticky top-24 space-y-6 text-sm"><DocNav title="开始" items={[['快速开始', '#quickstart'], ['鉴权', '#auth'], ['第一个请求', '#first-request']]} /><DocNav title="API" items={[['Responses', '#responses'], ['兼容端点', '#chat'], ['错误处理', '#errors']]} /><DocNav title="服务" items={[['条款与隐私', '#terms'], ['支持', '#support']]} /></div></aside>
        <article className="min-w-0 max-w-4xl">
          <div className="border-b border-border pb-10"><p className="text-xs font-bold uppercase tracking-[.2em] text-primary">Developer docs</p><h1 className="mt-3 text-4xl font-bold tracking-tight">快速开始</h1><p className="mt-4 text-lg leading-8 text-muted-foreground">用现有 OpenAI 或 Anthropic SDK 接入 AxisRelay。大多数项目只需要替换 Base URL 和 API Key。</p></div>
          <DocSection id="quickstart" icon={Rocket} title="三步完成接入">
            <div className="grid gap-4 sm:grid-cols-3">{[['1', '注册账号', '创建用户账号并进入控制台。'], ['2', '创建 Key', 'Key 只在创建时完整显示。'], ['3', '替换端点', '把 SDK Base URL 指向 /v1。']].map(([n, title, text]) => <div key={n} className="rounded-xl border border-border bg-card p-5"><div className="flex size-7 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground">{n}</div><h3 className="mt-4 font-semibold">{title}</h3><p className="mt-2 text-sm leading-6 text-muted-foreground">{text}</p></div>)}</div>
            <Button className="mt-6" asChild><Link to="/auth/register">创建免费账号</Link></Button>
          </DocSection>
          <DocSection id="auth" icon={KeyRound} title="鉴权">
            <p>所有 `/v1/*` 请求通过标准 Bearer Token 鉴权。不要把 API Key 放在浏览器前端、移动端安装包或公开仓库中。</p>
            <CodeBlock code={'Authorization: Bearer sk-axis-...'} />
          </DocSection>
          <DocSection id="first-request" icon={Terminal} title="第一个请求">
            <p>下面的请求使用 Responses API。模型名称取决于部署实例实际开放的模型。</p><CodeBlock code={curlExample} />
          </DocSection>
          <DocSection id="responses" icon={MessageSquareCode} title="OpenAI SDK">
            <p>官方 OpenAI SDK 可以直接使用。流式请求、工具调用和用量字段会按兼容协议返回。</p><CodeBlock code={jsExample} />
          </DocSection>
          <DocSection id="chat" icon={BookOpen} title="兼容端点">
            <div className="overflow-hidden rounded-xl border border-border"><table className="w-full text-sm"><tbody>{[['POST', '/v1/responses', 'Responses API'], ['POST', '/v1/chat/completions', 'Chat Completions'], ['POST', '/v1/messages', 'Anthropic Messages'], ['GET', '/v1/models', '可用模型列表'], ['POST', '/v1/images/generations', '图像生成']].map(([method, path, label]) => <tr key={path} className="border-b border-border last:border-0"><td className="w-24 px-4 py-3 font-mono text-xs text-primary">{method}</td><td className="px-4 py-3 font-mono text-xs">{path}</td><td className="px-4 py-3 text-muted-foreground">{label}</td></tr>)}</tbody></table></div>
          </DocSection>
          <DocSection id="errors" icon={Check} title="错误处理">
            <p>对 `429` 和 `5xx` 使用带抖动的指数退避；流式响应开始后，应把错误事件当作失败终态，不要把连接关闭自行解释为成功。</p>
            <div className="mt-5 rounded-xl border border-amber-500/20 bg-amber-500/8 p-4 text-sm"><strong>生产建议：</strong>为每个请求设置可取消的超时，并记录返回的请求 ID，便于在管理后台定位。</div>
          </DocSection>
          <DocSection id="terms" icon={BookOpen} title="服务条款与隐私说明">
            <div className="space-y-4">
              <p><strong className="text-foreground">账号与凭据：</strong>用户应妥善保管账号密码和 API Key，不得转售、共享或用于违法及未经授权的访问。发现泄露后应立即撤销 Key 或修改密码。</p>
              <p><strong className="text-foreground">合理使用：</strong>请求受实例管理员配置的额度、并发和内容安全策略约束。滥用、攻击、规避限制或干扰服务的账号可被暂停。</p>
              <p><strong className="text-foreground">数据处理：</strong>本实例保存注册邮箱、个人资料、登录设备信息、API Key 元数据与请求用量，以提供鉴权、计费、安全审计和故障排查。完整 API Key 仅在创建时展示；密码以不可逆哈希保存。</p>
              <p><strong className="text-foreground">模型请求：</strong>请求内容会发送至用户选择的上游模型提供方，并受相应提供方条款约束。请勿提交无权处理的敏感信息。</p>
              <p><strong className="text-foreground">可用性与责任：</strong>模型输出可能不准确，服务也可能因维护或上游故障暂时中断。生产使用方应自行设置超时、重试、审查和备份机制。</p>
              <p>继续注册或使用即表示接受上述规则。具体运营主体、保留期限、退款和争议条款由部署本实例的管理员另行公布；本页不替代其法定隐私政策。</p>
            </div>
          </DocSection>
          <section id="support" className="mt-14 rounded-2xl border border-border bg-card p-6"><h2 className="text-xl font-bold">需要部署或企业支持？</h2><p className="mt-2 text-sm text-muted-foreground">请联系当前 AxisRelay 实例的管理员。此站点不会把账号、密钥或支持请求发送给第三方。</p></section>
        </article>
      </div>
    </div>
  )
}

function DocNav({ title, items }: { title: string; items: Array<[string, string]> }) { return <div><div className="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">{title}</div><div className="grid gap-1">{items.map(([label, href]) => <a key={href} href={href} className="rounded-lg px-3 py-2 text-muted-foreground hover:bg-accent hover:text-foreground">{label}</a>)}</div></div> }

function DocSection({ id, icon: Icon, title, children }: { id: string; icon: typeof Rocket; title: string; children: React.ReactNode }) { return <section id={id} className="scroll-mt-24 border-b border-border py-10 last:border-0"><div className="mb-5 flex items-center gap-3"><span className="flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary"><Icon className="size-4" /></span><h2 className="text-2xl font-bold">{title}</h2></div><div className="text-sm leading-7 text-muted-foreground">{children}</div></section> }

function CodeBlock({ code }: { code: string }) {
  const [copied, setCopied] = useState(false)
  return <div className="relative mt-5 overflow-hidden rounded-xl border border-white/10 bg-[#0b1020]"><button className="absolute right-3 top-3 flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-2 py-1 text-xs text-slate-400 hover:text-white" onClick={() => { void navigator.clipboard.writeText(code); setCopied(true); window.setTimeout(() => setCopied(false), 1500) }}>{copied ? <Check className="size-3" /> : <Copy className="size-3" />}{copied ? '已复制' : '复制'}</button><pre className="overflow-x-auto p-5 pr-20 text-[13px] leading-6 text-slate-300"><code>{code}</code></pre></div>
}
