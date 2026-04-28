/** @type {import('tailwindcss').Config} */

const defaultTheme = require('tailwindcss/defaultTheme')

module.exports = {
  content: [
    'resources/views/**/*.jet',
    'resources/js/**/*.{vue,ts,tsx,js,jsx}',
  ],
  theme: {
    extend: {
      colors: {
        'pink-100': "#FBFBFB",
        'pink-200': "#EB4765",
        'pink-300': "#490814"
      },
      fontFamily: {
        'sans': ['Roboto', ...defaultTheme.fontFamily.sans]
      }
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
    require('@tailwindcss/forms'),
    require('@tailwindcss/aspect-ratio'),
    require('@tailwindcss/container-queries'),
  ],
}
