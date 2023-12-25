/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/*.{html,js,templ,go}"],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: { themes: false },
};
