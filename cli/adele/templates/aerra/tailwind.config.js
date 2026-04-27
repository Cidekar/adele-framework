/** @type {import("tailwindcss").Config} */

const defaultTheme = require("tailwindcss/defaultTheme")

module.exports = {
  content: [
    "resources/views/**/*.jet",
  ],
  theme: {
    extend: {
      animation: {
        "blink": "blink-animation 1s steps(5, start) infinite",
      },
      keyframes: {
        "blink-animation": {
          "to": { visibility: "hidden" }
        }
      },
      colors: {
        "pink-50": "#FDEDF0",
        "pink-100": "#FBFBFB",
        "pink-140": "#FBDAE0",
        "pink-145": "#F7B6C2",
        "pink-150": "#F9C8D1",
        "pink-200": "#EB4765",
        "pink-250": "#F5A3B2",
        "pink-300": "#F391A3",
        "pink-500": "#EB4765",
        "pink-550": "#E93556",
        "pink-1000": "#5C0A19",
        "pink-1150": "#25040A",
        "pink-1200": "#120205",
        "pink-1250": "#120205",
        "teal-50": "#effef9",
        "teal-100": "#cafdee",
        "teal-200": "#95fadf",
        "teal-300": "#58f0cd",
        "teal-400": "#26dbb7",
        "teal-500": "#0fd4b0",
        "teal-600": "#089981",
        "teal-700": "#0b7a69",
        "teal-800": "#0e6156",
        "teal-900": "#115047",
        "teal-950": "#02312d",
        "red-100": "#A5122D",
        "white" : "#FFFFFF",
      },
      fontFamily: {
        "sans": ["Roboto", ...defaultTheme.fontFamily.sans]
      }
    },
  },
  plugins: [
    require("@tailwindcss/typography"),
    require("@tailwindcss/forms"),
    require("@tailwindcss/aspect-ratio"),
    require("@tailwindcss/container-queries"),
  ],
}
