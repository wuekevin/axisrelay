import { useRoutes } from 'react-router-dom'
import { adminRoutes } from './admin.routes'

export default function AppRouter() {
  return useRoutes(adminRoutes)
}
