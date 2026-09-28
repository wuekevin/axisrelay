import { lazy } from 'react'
import { Navigate, type RouteObject } from 'react-router-dom'
import ConsoleLayout from '../layouts/ConsoleLayout'
import { ConsoleGuard } from './guards'
import { paths } from './paths'

const ConsoleDashboard = lazy(() => import('../pages/console/ConsoleDashboard'))
const ConsoleAPIKeys = lazy(() => import('../pages/console/ConsoleAPIKeys'))
const ConsoleUsage = lazy(() => import('../pages/console/ConsoleUsage'))
const ConsoleProfile = lazy(() => import('../pages/console/ConsoleProfile'))
const ConsoleSecurity = lazy(() => import('../pages/console/ConsoleSecurity'))

export const consoleRoutes: RouteObject[] = [
  {
    path: paths.console.root,
    element: <ConsoleGuard />,
    children: [
      {
        element: <ConsoleLayout />,
        children: [
          { index: true, element: <ConsoleDashboard /> },
          { path: 'keys', element: <ConsoleAPIKeys /> },
          { path: 'usage', element: <ConsoleUsage /> },
          { path: 'profile', element: <ConsoleProfile /> },
          { path: 'security', element: <ConsoleSecurity /> },
          { path: '*', element: <Navigate to={paths.console.root} replace /> },
        ],
      },
    ],
  },
]
