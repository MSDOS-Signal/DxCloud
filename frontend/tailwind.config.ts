import type { Config } from 'tailwindcss'

export default <Partial<Config>>{
  content: [
    './components/**/*.{vue,js,ts}',
    './layouts/**/*.vue',
    './pages/**/*.vue',
    './app.vue',
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          50: '#eff5ff',
          100: '#dbe7fe',
          200: '#bfdcff',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#006eff',
          600: '#0052d9',
          700: '#003aab',
          800: '#002a78',
          900: '#001a4d',
        },
      },
    },
  },
  plugins: [],
}
