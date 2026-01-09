import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
// export default defineConfig({
//   plugins: [react()],
// })


export default defineConfig({
  server: {
    proxy: {
      // Intercepts anything starting with "/api"
      '/api': {
        target: 'http://localhost:8080', // Your Go app port
        changeOrigin: true,
      },
      // Intercepts WebSocket connections
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
});
