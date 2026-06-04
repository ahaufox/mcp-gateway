import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { ThemeProvider } from './context/ThemeContext'
import { Navbar } from './components/Navbar'
import { Footer } from './components/Footer'
import { Changelog } from './pages/Changelog'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ThemeProvider>
      <div className="min-h-screen bg-gray-950 bg-gradient-mesh flex flex-col leading-relaxed text-gray-200">
        <Navbar activePage="changelog" />
        <main className="max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 flex-1">
          <Changelog />
        </main>
        <Footer />
      </div>
    </ThemeProvider>
  </StrictMode>,
)