import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { viteSingleFile } from 'vite-plugin-singlefile'

export default defineConfig({
  plugins: [react(), viteSingleFile()],
  build: {
    outDir: '../cmd/server',
    emptyOutDir: false,
  },
  server: {
    proxy: {
      '/api': 'http://localhost:7391',
      //  '/api': 'http://outbound.xsdhy.com:7391',
    },
  },
})
