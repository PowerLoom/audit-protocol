// this means if app restart {MAX_RESTART} times in 1 min then it stops
const MAX_RESTART = 10;
const MIN_UPTIME = 60000;

const NODE_ENV = "development" // process.env.NODE_ENV || 'development';
const CWD = "/home/ubuntu/audit-protocol-private/"

module.exports = {
  apps : [
    {
      name   : "ap-dag-finalizer",
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
      name   : "ap-diff-service",
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
      name   : "ap-backend",
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
      name   : "ap-payload-commit",
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
      name   : "ap-token-aggregator",
      script : "./uniswapTokenData",
      cwd : CWD+"token-aggregator",
      max_restarts: MAX_RESTART,
      env: {
        NODE_ENV: NODE_ENV,
      },
      args: "5"
    },
    {
      name   : "ap-proto-indexer",
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
      name   : "ap-dag-verifier",
      script : "./dagChainVerifier",
      cwd : CWD+"dag_verifier",
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      kill_timeout : 3000,
      env: {
        NODE_ENV: NODE_ENV,
      },
      args: "5" //Log level set to debug, for production change to 4 (INFO) or 2(ERROR)
    },
    {
      name   : "ap-pruning-archival-service",
      script : "./pruningArchivalService",
      cwd : CWD+"go-pruning-archival-service",
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
