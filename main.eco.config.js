module.exports = {
	apps: [

		{
			name: "main",
			script: "./gunicorn_main_launcher.py",
			watch: true
		},
		{
			name: "wl",
			script: "./gunicorn_webhook_launcher.py",
			watch: true
		},
		{
			name: "pc",
			script: "./payload_commit_service.py",
			watch: true
		},
	]
}
