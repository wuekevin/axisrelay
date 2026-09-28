import { useEffect, useState, type MouseEvent } from 'react'
import { BarChart3, BookOpen, Gauge, KeyRound, LogOut, Menu, Moon, ShieldCheck, Sun, UserRound, X } from 'lucide-react'
import { Link, NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { Button } from '../components/ui/button'
import { DEFAULT_SITE_NAME, useBranding } from '../branding'
import { useTheme } from '../hooks/useTheme'
import PortalLogo from '../portal/PortalLogo'
import { useUserAuth } from '../providers/AppProviders'

const items = [
  { to: '/console', label: '概览', icon: Gauge, end: true },
  { to: '/console/keys', label: 'API Key', icon: KeyRound },
  { to: '/console/usage', label: '用量中心', icon: BarChart3 },
  { to: '/console/profile', label: '个人资料', icon: UserRound },
  { to: '/console/security', label: '账号安全', icon: ShieldCheck },
]

export default function ConsoleLayout() {
  const auth = useUserAuth()
  const location = useLocation()
  const navigate = useNavigate()
  const { theme, toggle } = useTheme()
  const { siteName } = useBranding()
  const [mobileOpen, setMobileOpen] = useState(false)
  const current = items.find((item) => item.end ? location.pathname === item.to : location.pathname.startsWith(item.to))
  useEffect(() => {
    document.title = `${current?.label || '用户中心'} · ${siteName === DEFAULT_SITE_NAME ? 'AxisRelay' : siteName}`
  }, [current?.label, siteName])

  const logout = async () => {
    await auth.logout()
    navigate('/', { replace: true })
  }

  return (
    <div className="min-h-dvh bg-background text-foreground lg:grid lg:grid-cols-[248px_minmax(0,1fr)]">
      <aside className="hidden min-h-dvh border-r border-border bg-card/70 lg:sticky lg:top-0 lg:flex lg:h-dvh lg:flex-col">
        <div className="flex h-17 items-center border-b border-border px-5"><PortalLogo /></div>
        <nav className="flex-1 space-y-1 p-3" aria-label="用户中心导航">
          {items.map(({ to, label, icon: Icon, end }) => <NavLink key={to} to={to} end={end} className={({ isActive }) => `flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition ${isActive ? 'bg-primary text-primary-foreground shadow-sm' : 'text-muted-foreground hover:bg-accent hover:text-foreground'}`}><Icon className="size-4" />{label}</NavLink>)}
        </nav>
        <div className="border-t border-border p-3"><Link to="/docs" className="mb-2 flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm text-muted-foreground hover:bg-accent hover:text-foreground"><BookOpen className="size-4" />开发文档</Link><div className="rounded-xl border border-border bg-background p-3"><div className="truncate text-sm font-semibold">{auth.user?.display_name || '用户'}</div><div className="mt-0.5 truncate text-xs text-muted-foreground">{auth.user?.email}</div><Button variant="ghost" size="sm" className="mt-2 w-full justify-start text-muted-foreground" onClick={() => void logout()}><LogOut />退出登录</Button></div></div>
      </aside>

      <div className="min-w-0">
        <header className="sticky top-0 z-40 flex h-17 items-center border-b border-border bg-background/90 px-4 backdrop-blur-lg sm:px-6 lg:px-8">
          <Button variant="ghost" size="icon" className="mr-2 lg:hidden" onClick={() => setMobileOpen(true)}><Menu /></Button>
          <div><h1 className="text-base font-semibold">{current?.label || '用户中心'}</h1><p className="hidden text-xs text-muted-foreground sm:block">管理你的访问凭据、用量与账号安全</p></div>
          <div className="ml-auto flex items-center gap-2"><Button variant="ghost" size="icon" aria-label="切换主题" onClick={(event: MouseEvent<HTMLButtonElement>) => toggle(event)}>{theme === 'dark' ? <Sun /> : <Moon />}</Button><Button variant="outline" size="sm" asChild><Link to="/">返回官网</Link></Button></div>
        </header>
        <main className="mx-auto w-full max-w-7xl px-4 py-7 sm:px-6 lg:px-8"><Outlet /></main>
      </div>

      {mobileOpen && <div className="fixed inset-0 z-50 lg:hidden"><button className="absolute inset-0 bg-black/50" aria-label="关闭导航" onClick={() => setMobileOpen(false)} /><aside className="relative flex h-full w-[280px] flex-col bg-card shadow-2xl"><div className="flex h-17 items-center justify-between border-b border-border px-4"><PortalLogo /><Button variant="ghost" size="icon" aria-label="关闭导航" onClick={() => setMobileOpen(false)}><X /></Button></div><nav className="flex-1 space-y-1 p-3">{items.map(({ to, label, icon: Icon, end }) => <NavLink key={to} to={to} end={end} onClick={() => setMobileOpen(false)} className={({ isActive }) => `flex items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium ${isActive ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-accent'}`}><Icon className="size-4" />{label}</NavLink>)}</nav><div className="border-t border-border p-4"><button className="flex w-full items-center justify-between rounded-lg px-2 py-2 text-left" onClick={() => void logout()}><span><span className="block truncate text-sm font-semibold">{auth.user?.display_name}</span><span className="block truncate text-xs text-muted-foreground">{auth.user?.email}</span></span><LogOut className="size-4 text-muted-foreground" /></button></div></aside></div>}
    </div>
  )
}
