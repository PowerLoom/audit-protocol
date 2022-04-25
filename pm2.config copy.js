// this means if app restart {MAX_RESTART} times in 1 min then it stops
const MAX_RESTART = 10;
const MIN_UPTIME = 60000; 

const NODE_ENV = "development" // process.env.NODE_ENV || 'development';

module.exports = {
  apps : [
    {
      name   : "audit-protocol-payload-commit-parallel",
      script : "python3 ./payload_commit_service_parallel.py",
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      env: {
        NODE_ENV: NODE_ENV,
      }
    },
    {
      name   : "audit-protocol-webhooks",
      script : "python3 ./gunicorn_webhook_launcher.py",
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      env: {
        NODE_ENV: NODE_ENV,
      }
    },
    {
      name   : "audit-protocol-dag-verifier",
      script : "./dag_verifier/dagChainVerifier",
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      env: {
        NODE_ENV: NODE_ENV,
      },
      args: "5"
    }
  ]
}
