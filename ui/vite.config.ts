import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
// vite.config.ts

import fs from 'fs';
import path from 'path';


let plugin = {
  name: 'zlatko-api',
  configureServer(server) {
    server.middlewares.use((req, res, next) => {
      if (fs.existsSync(path.join('./vendor/', req.originalUrl))) {
        // fs.readFileSync ...
      } else {
        next();
      }
    });
  },
};


export default defineConfig({
  plugins: [vue()],
    server: {
      proxy: {
        '/assets/images/golang-back-empty.png': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          rewrite: (path) => path,
        },
        '/assets/images/golang-back.png': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          rewrite: (path) => path,
        },
        '/assets/images/back.png': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          rewrite: (path) => path,
        },
        '/assets/images/*.png': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          rewrite: (path) => path,
        },
        '/assets/pool/': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          rewrite: (path) => path,
        },
      },
  },
})
