/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ["./src/routes/**/*.{svelte,js}"],
	plugins: [require("daisyui")],
	daisyui: {
		themes: [
			{
				darkpink: {
					primary: "#fecdd3",
					secondary: "#7dd3fc",
					accent: "#fca5a5",
					neutral: "#44403c",
					"base-100": "#292524",
					info: "#93c5fd",
					success: "#10b981",
					warning: "#a16207",
					error: "#7f1d1d",
				},
			},
			{
				lightpink: {
					primary: "#fda4af",
					secondary: "#60a5fa",
					accent: "#f472b6",
					neutral: "#44403c",
					"base-100": "#f3f4f6",
					info: "#06b6d4",
					success: "#34d399",
					warning: "#f59e0b",
					error: "#dc2626",
				},
			},
		],
	},
}
