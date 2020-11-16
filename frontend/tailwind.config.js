module.exports = {
	future: {
		removeDeprecatedGapUtilities: true,
		purgeLayersByDefault: true,
	},
	purge: {
	enabled: true,
	content: ['./src/**/*.svelte'],
	},
	theme: {
		extend: {},
	},
	variants: {},
	plugins: [
			require('@tailwindcss/ui'),
	],
}
