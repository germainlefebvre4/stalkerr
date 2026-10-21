import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

const apiPort = process.env.API_PORT || 8080;
const frontendPort = Number(process.env.FRONTEND_PORT) || 5173;

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: frontendPort,
    strictPort: true,
    proxy: {
      '/api/v1': {
        target: `http://localhost:${apiPort}`,
        changeOrigin: true,
      },
      '/health': {
        target: `http://localhost:${apiPort}`,
        changeOrigin: true,
      },
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
  },
});
