import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import devApiMock from './mock/dev-api.js'

export default defineConfig({
  plugins: [devApiMock(), vue()],
  build: {
    outDir: '../frontend-build',
    emptyOutDir: true,
    sourcemap: true,
  },
})
