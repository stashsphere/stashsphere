import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { extractIcons } from './vite-plugin-icons';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [extractIcons(), react(), tailwindcss()],
});
