import { Suspense, type PropsWithChildren } from 'react'
import RouteErrorBoundary from './components/RouteErrorBoundary'
import StateShell from './components/StateShell'
import { ToastProvider } from './components/ToastProvider'
import { BrandingProvider } from './branding'
import { ThemeProvider } from './hooks/useTheme'
import AppRouter from './router'

export default function App() {
  return (
    <AppProviders>
      <AppRouter />
    </AppProviders>
  )
}

function AppProviders({ children }: PropsWithChildren) {
  return (
    <ThemeProvider>
      <BrandingProvider>
        <ToastProvider>
          <RouteErrorBoundary>
            <Suspense fallback={<StateShell variant="page" loading>{null}</StateShell>}>
              {children}
            </Suspense>
          </RouteErrorBoundary>
        </ToastProvider>
      </BrandingProvider>
    </ThemeProvider>
  )
}
