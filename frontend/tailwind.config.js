export default {
  content: ["./index.html", "./src/**/*.{js,jsx}"],
  theme: {
    extend: {
      fontFamily: {
        inter: ["Inter", "sans-serif"],
      },
      colors: {
        primary: "#38bdf8",
        secondary: "#0ea5e9",
        dark: "#0f172a",
      },
      boxShadow: {
        glass: "0 4px 30px rgba(0,0,0,0.4)",
      },
    },
  },
  plugins: [],
};
