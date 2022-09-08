// this means if app restart {MAX_RESTART} times in 1 min then it stops
const MAX_RESTART = 10;
const MIN_UPTIME = 60000; 

const NODE_ENV = "development" // process.env.NODE_ENV || 'development';
const CWD = "/home/ubuntu/audit-protocol-private/"

module.exports = {
  apps : [
    {
      name   : "audit-protocol-dag-finalizer",
      script : "python3 ./gunicorn_dag_finalizer_launcher.py",
      cwd : CWD,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      kill_timeout : 3000,
      env: {
        NODE_ENV: NODE_ENV,
        GUNICORN_WORKERS: 20
      }
    },
    {
      name   : "audit-protocol-diff-service",
      script : "python3 ./diff_calculation_service.py",
      cwd : CWD,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      kill_timeout : 3000,
      env: {
        NODE_ENV: NODE_ENV,
      }
    },
    {
      name   : "audit-protocol-backend",
      script : "python3 ./gunicorn_main_launcher.py",
      cwd : CWD,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      kill_timeout : 3000,
      env: {
        NODE_ENV: NODE_ENV,
        GUNICORN_WORKERS: 20
      },
    },
    {
      name   : "audit-protocol-payload-commit",
      script : "./payloadCommitService",
      cwd : CWD+"go-payload-commit-service",
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      kill_timeout : 3000,
      env: {
        NODE_ENV: NODE_ENV
      },
      args: "5" //Log level set to debug, for production change to 4 (INFO) or 2(ERROR)
    },
    {
      name   : "audit-protocol-proto-indexer",
      script : "python proto_sliding_window_cacher_service.py",
      cwd : CWD,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      kill_timeout : 3000,
      env: {
        NODE_ENV: NODE_ENV
      }
    },
    {
      name   : "audit-protocol-dag-verifier",
      script : "./dagChainVerifier",
      cwd : CWD+"dag_verifier",
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      kill_timeout : 3000,
      env: {
        NODE_ENV: NODE_ENV,
      },
      args: "5" //Log level set to debug, for production change to 4 (INFO) or 2(ERROR)
    }
  ]
}
