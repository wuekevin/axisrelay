import { Link } from 'react-router-dom'
import { DEFAULT_SITE_LOGO, useBranding } from '../branding'
import { cn } from '../lib/utils'

export default function PortalLogo({ compact = false, className }: { compact?: boolean; className?: string }) {
  const { siteLogo, siteName } = useBranding()
  return (
    <Link to="/" className={cn('inline-flex items-center gap-2.5 font-semibold tracking-tight text-foreground', className)}>
      <span className="relative flex size-9 items-center justify-center overflow-hidden rounded-xl border border-border/70 bg-card shadow-sm">
        <img src={siteLogo || DEFAULT_SITE_LOGO} alt="" className="size-full object-cover" />
      </span>
      {!compact && <span className="text-[17px]">{siteName === 'CodexProxy' ? 'AxisRelay' : siteName}</span>}
    </Link>
  )
}
