import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// During `npm run dev`, Vite proxies /api/* to the Go server on :8080.
// In production, the SPA is served separately (nginx / static host /
// behind the same reverse proxy as the Go server).
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
      '/healthz': 'http://localhost:8080',
      '/readyz': 'http://localhost:8080',
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    sourcemap: true,
  },
});
