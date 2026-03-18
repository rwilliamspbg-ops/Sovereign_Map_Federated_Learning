import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // Split Recharts into separate chunk to reduce main bundle
          recharts: ['recharts'],
          // React vendor chunk for better caching
          react: ['react', 'react-dom']
        }
      }
    },
    // Chunk size warning threshold
    chunkSizeWarningLimit: 600
  },
  server: {
    host: '0.0.0.0',
    port: 3000,
    strictPort: true,
    proxy: {
      '/backend': {
        target: 'http://localhost:8000',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/backend/, ''),
      },
      '/node-api': {
        target: 'http://localhost:8082',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/node-api/, ''),
      },
    },
  },
})
