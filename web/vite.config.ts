import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// The app is served by the Go server under /web/ (see web/fs.go).
export default defineConfig({
  base: '/web/',
  plugins: [react(), tailwindcss()],
  build: {
    outDir: 'dist',
    chunkSizeWarningLimit: 1200
  },
  server: {
    port: 5174,
    proxy: {
      '/api': { target: 'http://localhost:9999', changeOrigin: true },
      '/img': { target: 'http://localhost:9999', changeOrigin: true },
      '/imghm': { target: 'http://localhost:9999', changeOrigin: true },
      '/heresphere': { target: 'http://localhost:9999', changeOrigin: true },
      '/download': { target: 'http://localhost:9999', changeOrigin: true },
      '/ws': { target: 'ws://localhost:9999', ws: true }
    }
  }
})
