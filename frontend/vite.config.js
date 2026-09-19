import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: '../frontend-build',
    emptyOutDir: true,
    sourcemap: true,
  },
})
