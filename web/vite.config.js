import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

// 读取 .env / .env.local, 后端端口走 STAR_API_PORT, 生产 build 用相对路径
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiPort = env.STAR_API_PORT || '8181'
  const apiHost = env.STAR_API_HOST || 'localhost'
  return {
    plugins: [react()],
    base: env.VITE_BASE || '/',
    server: {
      host: '0.0.0.0',
      port: 5173,
      proxy: {
        '/api': {
          target: `http://${apiHost}:${apiPort}`,
          changeOrigin: true
        }
      }
    },
    build: {
      // 生产 build 跟 .env.production / .env 走
      outDir: 'dist',
      sourcemap: false,
      target: 'es2018',
      rollupOptions: {
        output: {
          manualChunks: {
            react: ['react', 'react-dom', 'react-router-dom']
          }
        }
      }
    }
  }
})