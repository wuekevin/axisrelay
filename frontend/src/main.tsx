import { StrictMode } from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import './i18n'
import './theme/arco.css'
import './index.css'

const rootElement = document.getElementById('root')
const routerBasename = window.location.pathname.startsWith('/admin') ? '/admin' : undefined

if (!rootElement) {
  throw new Error('未找到 root 节点')
}

ReactDOM.createRoot(rootElement).render(
  <StrictMode>
    <BrowserRouter basename={routerBasename}>
      <App />
    </BrowserRouter>
  </StrictMode>,
)
