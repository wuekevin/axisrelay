import { lazy } from 'react'
import { Navigate, type RouteObject } from 'react-router-dom'
import AdminLayout from '../layouts/AdminLayout'
import Dashboard from '../pages/Dashboard'
import { AdminGuard } from './guards'
import { paths } from './paths'

const Accounts = lazy(() => import('../pages/Accounts'))
const Operations = lazy(() => import('../pages/Operations'))
const OperationsErrors = lazy(() => import('../pages/OperationsErrors'))
const RuntimeStatus = lazy(() => import('../pages/RuntimeStatus'))
const Proxies = lazy(() => import('../pages/Proxies'))
const SchedulerBoard = lazy(() => import('../pages/SchedulerBoard'))
const Settings = lazy(() => import('../pages/Settings'))
const Docs = lazy(() => import('../pages/Docs'))
const APIKeys = lazy(() => import('../pages/APIKeys'))
const Usage = lazy(() => import('../pages/Usage'))
const ImageStudio = lazy(() => import('../pages/ImageStudio'))
const QualityTest = lazy(() => import('../pages/QualityTest'))
const PromptFilter = lazy(() => import('../pages/PromptFilter'))
const ThemeSettings = lazy(() => import('../pages/ThemeSettings'))
const ModelPricing = lazy(() => import('../pages/ModelPricing'))
const PayloadRules = lazy(() => import('../pages/PayloadRules'))

export const adminRoutes: RouteObject[] = [
  {
    path: paths.admin.root,
    element: <AdminGuard />,
    children: [
      {
        element: <AdminLayout />,
        children: [
          { index: true, element: <Navigate to={paths.admin.gateway.root} replace /> },
          { path: 'gateway', element: <Dashboard /> },
          { path: 'gateway/accounts', element: <Accounts /> },
          { path: 'gateway/accounts/grok', element: <Accounts /> },
          { path: 'gateway/accounts/antigravity', element: <Accounts /> },
          { path: 'gateway/accounts/claude', element: <Accounts /> },
          { path: 'gateway/accounts/invite', element: <Accounts /> },
          { path: 'gateway/api-keys', element: <APIKeys /> },
          { path: 'gateway/proxies', element: <Proxies /> },
          { path: 'gateway/images', element: <Navigate to={paths.admin.gateway.imageStudio} replace /> },
          { path: 'gateway/images/:view', element: <ImageStudio /> },
          { path: 'gateway/quality-test', element: <QualityTest /> },
          { path: 'gateway/prompt-filter', element: <Navigate to={paths.admin.gateway.promptFilterOverview} replace /> },
          { path: 'gateway/prompt-filter/:view', element: <PromptFilter /> },
          { path: 'gateway/ops', element: <Navigate to={paths.admin.gateway.operationsOverview} replace /> },
          { path: 'gateway/ops/overview', element: <Operations /> },
          { path: 'gateway/ops/runtime', element: <RuntimeStatus /> },
          { path: 'gateway/ops/errors', element: <OperationsErrors /> },
          { path: 'gateway/ops/scheduler', element: <SchedulerBoard /> },
          { path: 'gateway/usage', element: <Usage /> },
          { path: 'gateway/model-pricing', element: <ModelPricing /> },
          { path: 'gateway/payload-rules', element: <Navigate to={paths.admin.gateway.payloadRulesEditor} replace /> },
          { path: 'gateway/payload-rules/:view', element: <PayloadRules /> },
          { path: 'gateway/theme', element: <ThemeSettings /> },
          { path: 'gateway/settings', element: <Settings /> },
          { path: 'gateway/docs', element: <Docs /> },
          { path: 'gateway/guide', element: <Navigate to={paths.admin.gateway.docs} replace /> },
          { path: 'gateway/api-reference', element: <Navigate to={`${paths.admin.gateway.docs}#model-api`} replace /> },
        ],
      },
    ],
  },
]
