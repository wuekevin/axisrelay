import { Suspense, type PropsWithChildren } from 'react'
import { ConfigProvider } from '@arco-design/web-react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { I18nextProvider } from 'react-i18next'
import i18n from '../i18n'
import { BrandingProvider } from '../branding'
import RouteErrorBoundary from '../components/RouteErrorBoundary'
import StateShell from '../components/StateShell'
import { ToastProvider } from '../components/ToastProvider'
import { ThemeProvider } from '../hooks/useTheme'

const queryClient = new QueryClient()

export default function AppProviders({ children }: PropsWithChildren) {
  return (
    <I18nextProvider i18n={i18n}>
      <ConfigProvider>
        <ThemeProvider>
          <QueryClientProvider client={queryClient}>
            <BrandingProvider>
              <ToastProvider>
                <RouteErrorBoundary>
                  <Suspense fallback={<StateShell variant="page" loading>{null}</StateShell>}>
                    {children}
                  </Suspense>
                </RouteErrorBoundary>
              </ToastProvider>
            </BrandingProvider>
          </QueryClientProvider>
        </ThemeProvider>
      </ConfigProvider>
    </I18nextProvider>
  )
}
