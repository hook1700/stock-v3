import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '/api')
      }
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    rollupOptions: {
      output: {
        // 静态资源分包
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
        // 手动分包：将第三方依赖单独打包
        manualChunks(id) {
          // 将 node_modules 中的代码分包
          if (id.includes('node_modules')) {
            // Vue 相关
            if (id.includes('vue') || id.includes('@vue')) {
              return 'vue-vendor'
            }
            // Element Plus
            if (id.includes('element-plus')) {
              return 'element-vendor'
            }
            // ECharts
            if (id.includes('echarts')) {
              return 'echarts-vendor'
            }
            // 其他第三方库打包在一起
            return 'vendor'
          }
        }
      }
    },
    // 增大 chunk 大小警告阈值到 1000KB
    chunkSizeWarningLimit: 1000
  }
})