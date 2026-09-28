import { useEffect, useState, type MouseEvent } from 'react'
import { ArrowRight, Menu, Moon, Sun, X } from 'lucide-react'
import { Link, NavLink, Outlet, useLocation } from 'react-router-dom'
import { DEFAULT_SITE_NAME, useBranding } from '../branding'
import { Button } from '../components/ui/button'
import { useTheme } from '../hooks/useTheme'
import PortalLogo from '../portal/PortalLogo'
import { useUserAuth } from '../providers/AppProviders'

const navItems = [
  { to: '/', label: '首页' },
  { to: '/models', label: '模型' },
  { to: '/pricing', label: '价格' },
  { to: '/docs', label: '文档' },
  { to: '/status', label: '状态' },
]

export default function PublicLayout() {
  const location = useLocation()
  const auth = useUserAuth()
  const { theme, toggle } = useTheme()
  const { siteName } = useBranding()
  const publicSiteName = siteName === DEFAULT_SITE_NAME ? 'AxisRelay' : siteName
  const [menuOpen, setMenuOpen] = useState(false)
  const standalone = ['/key-usage', '/image-studio', '/account-portal'].some((path) => location.pathname.startsWith(path))

  useEffect(() => {
    setMenuOpen(false)
    const label = navItems.find((item) => item.to === location.pathname)?.label
    document.title = label && label !== '首页' ? `${label} · ${publicSiteName}` : `${publicSiteName} · 统一 AI 模型网关`
  }, [location.pathname, publicSiteName])

  if (standalone) return <Outlet />

  return (
    <div className="min-h-dvh bg-background text-foreground">
      <header className="sticky top-0 z-50 border-b border-border/70 bg-background/85 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
          <PortalLogo />
          <nav className="hidden items-center gap-1 md:flex" aria-label="主导航">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.to === '/'}
                className={({ isActive }) => `rounded-lg px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-accent text-foreground' : 'text-muted-foreground hover:bg-accent/70 hover:text-foreground'}`}
              >
                {item.label}
              </NavLink>
            ))}
          </nav>
          <div className="hidden items-center gap-2 md:flex">
            <Button variant="ghost" size="icon" aria-label="切换深浅主题" onClick={(event: MouseEvent<HTMLButtonElement>) => toggle(event)}>
              {theme === 'dark' ? <Sun /> : <Moon />}
            </Button>
            {auth.authenticated ? (
              <Button asChild><Link to="/console">进入用户中心 <ArrowRight /></Link></Button>
            ) : (
              <>
                <Button variant="ghost" asChild><Link to="/auth/login">登录</Link></Button>
                <Button asChild><Link to="/auth/register">免费注册 <ArrowRight /></Link></Button>
              </>
            )}
          </div>
          <Button variant="ghost" size="icon" className="md:hidden" aria-label="打开导航" onClick={() => setMenuOpen((value) => !value)}>
            {menuOpen ? <X /> : <Menu />}
          </Button>
        </div>
        {menuOpen && (
          <div className="border-t border-border bg-background px-4 py-4 md:hidden">
            <nav className="mx-auto grid max-w-7xl gap-1">
              {navItems.map((item) => (
                <NavLink key={item.to} to={item.to} end={item.to === '/'} className="rounded-lg px-3 py-2.5 text-sm font-medium hover:bg-accent">
                  {item.label}
                </NavLink>
              ))}
              <div className="mt-3 grid grid-cols-2 gap-2 border-t border-border pt-4">
                <Button variant="outline" asChild><Link to={auth.authenticated ? '/console' : '/auth/login'}>{auth.authenticated ? '用户中心' : '登录'}</Link></Button>
                <Button asChild><Link to={auth.authenticated ? '/console/keys' : '/auth/register'}>{auth.authenticated ? 'API Key' : '免费注册'}</Link></Button>
              </div>
            </nav>
          </div>
        )}
      </header>

      <main><Outlet /></main>

      <footer className="border-t border-border bg-card/40">
        <div className="mx-auto grid max-w-7xl gap-10 px-4 py-12 sm:px-6 md:grid-cols-[1.4fr_1fr_1fr_1fr] lg:px-8">
          <div>
            <PortalLogo />
            <p className="mt-4 max-w-sm text-sm leading-6 text-muted-foreground">一个端点连接多种主流 AI 模型。统一协议、稳定调度、透明用量，让团队把精力放在产品而不是网关维护上。</p>
          </div>
          <FooterColumn title="产品" items={[['模型', '/models'], ['价格', '/pricing'], ['服务状态', '/status']]} />
          <FooterColumn title="开发者" items={[['快速开始', '/docs'], ['API 文档', '/docs#responses'], ['用户中心', '/console']]} />
          <FooterColumn title="服务" items={[['登录', '/auth/login'], ['注册', '/auth/register'], ['管理后台', '/admin/']]} />
        </div>
        <div className="border-t border-border/70 px-4 py-5 text-center text-xs text-muted-foreground">© {new Date().getFullYear()} AxisRelay. 面向可靠 AI 基础设施而构建。</div>
      </footer>
    </div>
  )
}

function FooterColumn({ title, items }: { title: string; items: Array<[string, string]> }) {
  return (
    <div>
      <h2 className="text-sm font-semibold">{title}</h2>
      <div className="mt-3 grid gap-2.5 text-sm text-muted-foreground">
        {items.map(([label, to]) => <Link key={to} to={to} className="w-fit hover:text-foreground">{label}</Link>)}
      </div>
    </div>
  )
}
