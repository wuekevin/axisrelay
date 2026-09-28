import type { PropsWithChildren } from 'react'
import { Outlet } from 'react-router-dom'
import LegacyAdminLayout from '../components/Layout'

export default function AdminLayout({ children }: PropsWithChildren) {
  return <LegacyAdminLayout>{children ?? <Outlet />}</LegacyAdminLayout>
}
