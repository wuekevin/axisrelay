import { lazy } from 'react'
import { Navigate, type RouteObject } from 'react-router-dom'
import PublicLayout from '../layouts/PublicLayout'
import { paths } from './paths'

const APIKeyUsagePortal = lazy(() => import('../pages/APIKeyUsagePortal'))
const ImageStudioPortal = lazy(() => import('../pages/ImageStudioPortal'))
const AccountPortal = lazy(() => import('../pages/AccountPortal'))
const PublicHome = lazy(() => import('../pages/PublicHome'))
const PublicModels = lazy(() => import('../pages/PublicModels'))
const PublicPricing = lazy(() => import('../pages/PublicPricing'))
const PublicDocs = lazy(() => import('../pages/PublicDocs'))
const PublicStatus = lazy(() => import('../pages/PublicStatus'))

export const publicRoutes: RouteObject[] = [
  {
    element: <PublicLayout />,
    children: [
      { path: paths.public.home, element: <PublicHome /> },
      { path: paths.public.models, element: <PublicModels /> },
      { path: paths.public.pricing, element: <PublicPricing /> },
      { path: paths.public.docs, element: <PublicDocs /> },
      { path: paths.public.status, element: <PublicStatus /> },
      { path: paths.public.keyUsage, element: <Navigate to="/key-usage/overview" replace /> },
      { path: '/key-usage/:view', element: <APIKeyUsagePortal /> },
      { path: paths.public.imageStudio, element: <Navigate to="/image-studio/studio" replace /> },
      { path: '/image-studio/:view', element: <ImageStudioPortal /> },
      { path: paths.public.accountPortal, element: <Navigate to="/account-portal/submit" replace /> },
      { path: '/account-portal/:view', element: <AccountPortal /> },
    ],
  },
]
