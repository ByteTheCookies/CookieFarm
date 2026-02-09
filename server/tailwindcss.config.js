const colors = require("tailwindcss/colors");

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./internal/ui/views/**/*.html", "./internal/ui/views/partials/**/*.html", "./internal/ui/views/layouts/**/*.html"],
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};
