/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: [
    "./pages/**/*.{js,ts,jsx,tsx}",
    "./components/**/*.{js,ts,jsx,tsx}"
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: "#2563eb",
          accent: "#0ea5e9"
        },
        light: {
          bg: "#ffffff",
          card: "#f9f9f9",
          text: "#111111",
        },
        dark: {
          bg: "#000000",
          card: "#111111",
          text: "#f9f9f9",
        }
      },
      boxShadow: {
        soft: "0 8px 24px rgba(17,24,39,0.08)"
      },
      borderRadius: {
        xl: "1rem",
        "2xl": "1.25rem"
      },
      transitionProperty: {
        'colors-transform': 'color, background-color, border-color, transform'
      }
    },
  },
  plugins: [],
}
