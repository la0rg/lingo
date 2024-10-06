/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/src/templates/**/*.html"],
  theme: {
    fontFamily: {
      sans: ['"Open Sans"', 'ui-sans-serif', 'system-ui', 'sans-serif', '"Apple Color Emoji"', '"Segoe UI Emoji"', '"Segoe UI Symbol"', '"Noto Color Emoji"'],
    },
    extend: {
      fontFamily: {
        title: ['"Raleway"', 'sans-serif', 'system-ui'], 
      },
    },
  },
  plugins: [
    require('daisyui'),
  ],
  daisyui: {
    themes: ["dark"],
  },
}