import {
  createContext,
  Suspense,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from 'react'
import { I18nextProvider } from 'react-i18next'
import i18n from '../i18n'
import { BrandingProvider } from '../branding'
import RouteErrorBoundary from '../components/RouteErrorBoundary'
import StateShell from '../components/StateShell'
import { ToastProvider } from '../components/ToastProvider'
import { ThemeProvider } from '../hooks/useTheme'
import { portalAPI } from '../portal/api'
import type { AuthSessionResponse, PortalUser } from '../portal/types'

type UserAuthContextValue = {
  status: 'loading' | 'authenticated' | 'anonymous'
  authenticated: boolean
  user: PortalUser | null
  setSession: (session: AuthSessionResponse) => void
  refresh: () => Promise<void>
  logout: () => Promise<void>
}

const UserAuthContext = createContext<UserAuthContextValue | null>(null)

export function useUserAuth() {
  const context = useContext(UserAuthContext)
  if (!context) {
    throw new Error('useUserAuth must be used inside <AppProviders>')
  }
  return context
}

function UserAuthProvider({ children }: PropsWithChildren) {
  const [status, setStatus] = useState<UserAuthContextValue['status']>('loading')
  const [user, setUser] = useState<PortalUser | null>(null)

  const setSession = useCallback((session: AuthSessionResponse) => {
    if (session.authenticated && session.user) {
      setUser(session.user)
      setStatus('authenticated')
      return
    }
    setUser(null)
    setStatus('anonymous')
  }, [])

  const refresh = useCallback(async () => {
    try {
      setSession(await portalAPI.session())
    } catch {
      setSession({ authenticated: false })
    }
  }, [setSession])

  const logout = useCallback(async () => {
    try {
      await portalAPI.logout()
    } finally {
      setSession({ authenticated: false })
    }
  }, [setSession])

  useEffect(() => {
    void refresh()
  }, [refresh])

  const value = useMemo(
    () => ({ status, authenticated: status === 'authenticated', user, setSession, refresh, logout }),
    [logout, refresh, setSession, status, user],
  )

  return <UserAuthContext.Provider value={value}>{children}</UserAuthContext.Provider>
}

export default function AppProviders({ children }: PropsWithChildren) {
  return (
    <I18nextProvider i18n={i18n}>
      <ThemeProvider>
        <UserAuthProvider>
          <BrandingProvider>
            <ToastProvider>
              <RouteErrorBoundary>
                <Suspense fallback={<StateShell variant="page" loading>{null}</StateShell>}>
                  {children}
                </Suspense>
              </RouteErrorBoundary>
            </ToastProvider>
          </BrandingProvider>
        </UserAuthProvider>
      </ThemeProvider>
    </I18nextProvider>
  )
}
