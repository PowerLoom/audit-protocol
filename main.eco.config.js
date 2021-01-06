module.exports = {
	apps: [

		{
			name: "new_main",
			script: "gunicorn_main_launcher.py",
			watch: true
		},
		{
			name: "new_wl",
			script: "gunicorn_webhook_launcher.py",
			watch: true
		},
	]
}
