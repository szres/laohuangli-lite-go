/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ['./src/routes/**/*.{svelte,js}'],
	plugins: [require("daisyui")],
	daisyui: {
	  themes: ["business", "valentine"],
	},
}
