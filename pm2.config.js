// this means if app restart {MAX_RESTART} times in 1 min then it stops
const MAX_RESTART = 10;
const MIN_UPTIME = 60000; 

const NODE_ENV = "development" // process.env.NODE_ENV || 'development';
const CWD = "/home/ubuntu/audit-protocol-private/"

module.exports = {
  apps : [
    {
      name   : "audit-protocol-backend",
      script : "python3 ./gunicorn_main_launcher.py",
      cwd : CWD,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      env: {
        NODE_ENV: NODE_ENV
      },
    },
    {
      name   : "audit-protocol-payload-commit",
      script : "./go-payload-commit-service/payloadCommitService",
      cwd : CWD,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      env: {
        NODE_ENV: NODE_ENV
      },
      args: "5" //Log level set to debug, for production change to 4 (INFO) or 2(ERROR)
    },
    {
      name   : "audit-protocol-webhooks",
      script : "python3 ./gunicorn_webhook_launcher.py",
      cwd : CWD,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      env: {
        NODE_ENV: NODE_ENV,
      }
    },
    {
      name   : "audit-protocol-dag-verifier",
      script : "./dag_verifier/dagChainVerifier",
      cwd : CWD,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      env: {
        NODE_ENV: NODE_ENV,
      },
      args: "5" //Log level set to debug, for production change to 4 (INFO) or 2(ERROR)
    }
  ]
}
