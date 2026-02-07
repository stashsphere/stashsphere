import { addIconSelectors } from '@iconify/tailwind';

/** @type {import('tailwindcss').Config} */
export default {
  theme: {
    extend: {
      rotate: {
        270: '270deg',
      },
    },
  },
  plugins: [addIconSelectors(['mdi'])],
  darkMode: 'selector',
};
