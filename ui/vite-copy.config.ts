import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
// vite.config.ts

export default defineConfig({
  plugins: [vue()],
    server: {
      proxy: {
        '/assets/images/golang-back-empty.png': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          rewrite: (path) => path,
        },
        '/assets/images/back.png': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          rewrite: (path) => path,
        },
      },
  },
})
