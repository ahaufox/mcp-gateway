import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:9090',
        changeOrigin: true,
      }
    }
  },
  build: {
    outDir: '../mcp-proxy/internal/server/frontend/dist',
    emptyOutDir: true,
    rollupOptions: {
      input: {
        dashboard: resolve(__dirname, 'dashboard.html'),
        converter: resolve(__dirname, 'converter.html'),
        changelog: resolve(__dirname, 'changelog.html'),
        login: resolve(__dirname, 'login.html'),
      }
    }
  }
})