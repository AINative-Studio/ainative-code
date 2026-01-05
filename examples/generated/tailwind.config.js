/** @type {import('tailwindcss').Config} */
module.exports = {
  theme: {
    extend: {
      colors: {
        "primary-color": "#007bff",
        "secondary-color": "#6c757d",
        "success-color": "#28a745",
      },
      spacing: {
        "spacing-base": "16px",
        "spacing-sm": "8px",
        "spacing-lg": "32px",
      },
      fontFamily: {
        "font-family-base": ['Helvetica', 'Arial', 'sans-serif'],
      },
      fontSize: {
        "font-size-base": "16px",
      },
      borderRadius: {
        "border-radius-sm": "4px",
      },
      boxShadow: {
        "shadow-sm": "0 1px 2px rgba(0,0,0,0.1)",
      },
    },
  },
  plugins: [],
}
