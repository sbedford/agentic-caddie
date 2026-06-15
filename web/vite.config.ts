import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '^/(players|courses|tees|course-holes|tee-holes|pois|clubs|rounds|holes|shots|commentary|vocabulary)': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
    },
  },
})
