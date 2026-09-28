import { lazy } from 'react'
import { Navigate, type RouteObject } from 'react-router-dom'
import AuthLayout from '../layouts/AuthLayout'
import { paths } from './paths'

const AuthPage = lazy(() => import('../pages/AuthPage'))

export const authRoutes: RouteObject[] = [
  {
    path: paths.auth.root,
    element: <AuthLayout />,
    children: [
      { index: true, element: <Navigate to="login" replace /> },
      { path: ':mode', element: <AuthPage /> },
      { path: '*', element: <Navigate to="login" replace /> },
    ],
  },
]
