import starlightPlugin from "@astrojs/starlight-tailwind";

// Generated color palettes
// https://starlight.astro.build/guides/css-and-tailwind/#color-theme-editor
const accent = { 200: '#D9BEF3', 600: '#853BCE', 900: '#3D2259', 950: '#291839' };
const gray = { 100: '#F7F7F8', 200: '#F1F1F3', 300: '#C4C4CA', 400: '#868593', 500: '#535260', 700: '#33323E', 800: '#201F2D', 900: '#1C1A28' };

/** @type {import('tailwindcss').Config} */
export default {
  content: ["./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}"],
  plugins: [starlightPlugin()],

  theme: {
    extend: {
      colors: { accent, gray },
      fontFamily: {
        sans: ["Inter", "ui-sans-serif", "sans-serif"],
      },
    },
  },
};
