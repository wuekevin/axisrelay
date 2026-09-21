import { lazy, Suspense } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import AuthGate from './components/AuthGate'
import RouteErrorBoundary from './components/RouteErrorBoundary'
import StateShell from './components/StateShell'
import { ToastProvider } from './components/ToastProvider'
import { BrandingProvider } from './branding'
import { VisibleChannelsProvider } from './visibleChannels'
import { ThemeProvider } from './hooks/useTheme'
import AdminLayout from './layouts/AdminLayout'
import AdminRoutes from './router/AdminRoutes'

const APIKeyUsagePortal = lazy(() => import('./pages/APIKeyUsagePortal'))
const ImageStudioPortal = lazy(() => import('./pages/ImageStudioPortal'))
const AccountPortal = lazy(() => import('./pages/AccountPortal'))

export default function App() {
  return (
    <ThemeProvider>
      <BrandingProvider>
        <ToastProvider>
          <RouteErrorBoundary>
            <Suspense fallback={<StateShell variant="page" loading>{null}</StateShell>}>
              <Routes>
                <Route path="/key-usage" element={<Navigate to="/key-usage/overview" replace />} />
                <Route path="/key-usage/:view" element={<APIKeyUsagePortal />} />
                <Route path="/image-studio" element={<Navigate to="/image-studio/studio" replace />} />
                <Route path="/image-studio/:view" element={<ImageStudioPortal />} />
                <Route path="/account-portal" element={<Navigate to="/account-portal/submit" replace />} />
                <Route path="/account-portal/:view" element={<AccountPortal />} />
                <Route path="/*" element={<AdminApp />} />
              </Routes>
            </Suspense>
          </RouteErrorBoundary>
        </ToastProvider>
      </BrandingProvider>
    </ThemeProvider>
  )
}

function AdminApp() {
  return (
    <AuthGate>
      <VisibleChannelsProvider>
        <AdminLayout>
          <AdminRoutes />
        </AdminLayout>
      </VisibleChannelsProvider>
    </AuthGate>
  )
}
