import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// Vite configuration for the React + Tailwind project.
// The React plugin adds JSX support and Fast Refresh.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    open: false,
  },
});
