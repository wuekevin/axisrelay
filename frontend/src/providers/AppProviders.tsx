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
import { ConfigProvider } from '@arco-design/web-react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { I18nextProvider } from 'react-i18next'
import i18n from '../i18n'
import { BrandingProvider } from '../branding'
import RouteErrorBoundary from '../components/RouteErrorBoundary'
import StateShell from '../components/StateShell'
import { ToastProvider } from '../components/ToastProvider'
import { ThemeProvider } from '../hooks/useTheme'
import { portalAPI } from '../portal/api'
import type { AuthSessionResponse, PortalUser } from '../portal/types'
import {
  ADMIN_AUTH_CHANGED_EVENT,
  ADMIN_AUTH_REQUIRED_EVENT,
  getAdminKey,
  resetAdminAuthState,
  setAdminKey,
} from '../api'

const queryClient = new QueryClient()

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

type AdminAuthContextValue = {
  adminKey: string
  authenticated: boolean
  setKey: (key: string) => void
  reset: () => void
}

const AdminAuthContext = createContext<AdminAuthContextValue | null>(null)

export function useAdminAuth() {
  const context = useContext(AdminAuthContext)
  if (!context) {
    throw new Error('useAdminAuth must be used inside <AppProviders>')
  }
  return context
}

function AdminAuthProvider({ children }: PropsWithChildren) {
  const [adminKey, setAdminKeyState] = useState(() => getAdminKey())

  useEffect(() => {
    const refresh = () => setAdminKeyState(getAdminKey())
    const handleStorage = (event: StorageEvent) => {
      if (event.key === 'admin_key' || event.key === 'admin_auth_reset_at') {
        refresh()
      }
    }

    window.addEventListener(ADMIN_AUTH_CHANGED_EVENT, refresh)
    window.addEventListener(ADMIN_AUTH_REQUIRED_EVENT, refresh)
    window.addEventListener('storage', handleStorage)

    return () => {
      window.removeEventListener(ADMIN_AUTH_CHANGED_EVENT, refresh)
      window.removeEventListener(ADMIN_AUTH_REQUIRED_EVENT, refresh)
      window.removeEventListener('storage', handleStorage)
    }
  }, [])

  const setKey = useCallback((key: string) => {
    setAdminKey(key)
  }, [])

  const reset = useCallback(() => {
    resetAdminAuthState()
  }, [])

  const value = useMemo(
    () => ({
      adminKey,
      authenticated: Boolean(adminKey),
      setKey,
      reset,
    }),
    [adminKey, reset, setKey],
  )

  return <AdminAuthContext.Provider value={value}>{children}</AdminAuthContext.Provider>
}

export default function AppProviders({ children }: PropsWithChildren) {
  return (
    <I18nextProvider i18n={i18n}>
      <ConfigProvider>
        <ThemeProvider>
          <QueryClientProvider client={queryClient}>
            <UserAuthProvider>
              <AdminAuthProvider>
                <BrandingProvider>
                  <ToastProvider>
                    <RouteErrorBoundary>
                      <Suspense fallback={<StateShell variant="page" loading>{null}</StateShell>}>
                        {children}
                      </Suspense>
                    </RouteErrorBoundary>
                  </ToastProvider>
                </BrandingProvider>
              </AdminAuthProvider>
            </UserAuthProvider>
          </QueryClientProvider>
        </ThemeProvider>
      </ConfigProvider>
    </I18nextProvider>
  )
}
