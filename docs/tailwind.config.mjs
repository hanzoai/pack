import starlightPlugin from "@astrojs/starlight-tailwind";

// Generated color palettes
// https://starlight.astro.build/guides/css-and-tailwind/#color-theme-editor
const accent = {
  200: "#f0b0e9",
  600: "#980093",
  900: "#570054",
  950: "#3e073b",
};
const gray = {
  100: "#f6f6f6",
  200: "#eeeeee",
  300: "#c2c2c2",
  400: "#8b8b8b",
  500: "#585858",
  700: "#383838",
  800: "#272727",
  900: "#181818",
};

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
