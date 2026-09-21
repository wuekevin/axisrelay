import type { RouteObject } from 'react-router-dom'
import AuthLayout from '../layouts/AuthLayout'
import { paths } from './paths'

export const authRoutes: RouteObject[] = [
  {
    path: `${paths.auth.root}/*`,
    element: <AuthLayout />,
  },
]
