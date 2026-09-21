import { lazy } from 'react'
import { Navigate, type RouteObject } from 'react-router-dom'
import PublicLayout from '../layouts/PublicLayout'
import { paths } from './paths'

const APIKeyUsagePortal = lazy(() => import('../pages/APIKeyUsagePortal'))
const ImageStudioPortal = lazy(() => import('../pages/ImageStudioPortal'))
const AccountPortal = lazy(() => import('../pages/AccountPortal'))

export const publicRoutes: RouteObject[] = [
  {
    element: <PublicLayout />,
    children: [
      { path: paths.public.home },
      { path: paths.public.models },
      { path: paths.public.pricing },
      { path: paths.public.docs },
      { path: paths.public.status },
      { path: paths.public.keyUsage, element: <Navigate to="/key-usage/overview" replace /> },
      { path: '/key-usage/:view', element: <APIKeyUsagePortal /> },
      { path: paths.public.imageStudio, element: <Navigate to="/image-studio/studio" replace /> },
      { path: '/image-studio/:view', element: <ImageStudioPortal /> },
      { path: paths.public.accountPortal, element: <Navigate to="/account-portal/submit" replace /> },
      { path: '/account-portal/:view', element: <AccountPortal /> },
    ],
  },
]
