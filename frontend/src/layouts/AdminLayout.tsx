import type { PropsWithChildren } from 'react'
import { ConfigProvider } from '@arco-design/web-react'
import { Outlet } from 'react-router-dom'
import LegacyAdminLayout from '../components/Layout'

export default function AdminLayout({ children }: PropsWithChildren) {
  return (
    <ConfigProvider>
      <LegacyAdminLayout>{children ?? <Outlet />}</LegacyAdminLayout>
    </ConfigProvider>
  )
}
