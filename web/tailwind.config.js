/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/src/templates/**/*.html"],
  theme: {
    extend: {},
  },
  plugins: [
    require('daisyui'),
  ],
  daisyui: {
    themes: ["dark"],
  },
}