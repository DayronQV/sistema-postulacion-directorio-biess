import { BrowserRouter } from 'react-router-dom'
import ContenedorToasts from '@/components/ui/ContenedorToasts'
import AppRoutes from '@/routes'

export default function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
      <ContenedorToasts />
    </BrowserRouter>
  )
}
