import type { RouteObject } from 'react-router-dom'
import ConsoleLayout from '../layouts/ConsoleLayout'
import { ConsoleGuard } from './guards'
import { paths } from './paths'

export const consoleRoutes: RouteObject[] = [
  {
    path: paths.console.root,
    element: <ConsoleGuard />,
    children: [
      {
        path: '*',
        element: <ConsoleLayout />,
      },
    ],
  },
]
