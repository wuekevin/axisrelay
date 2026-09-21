import type { PropsWithChildren } from 'react'
import { ConfigProvider } from '@arco-design/web-react'
import LegacyAdminLayout from '../components/Layout'

export default function AdminLayout({ children }: PropsWithChildren) {
  return (
    <ConfigProvider>
      <LegacyAdminLayout>{children}</LegacyAdminLayout>
    </ConfigProvider>
  )
}
