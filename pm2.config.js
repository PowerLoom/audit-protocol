// this means if app restart {MAX_RESTART} times in 1 min then it stops
const MAX_RESTART = 10;
const MIN_UPTIME = 60000;
const NODE_ENV = process.env.NODE_ENV || 'development';
const INTERPRETER = process.env.AP_INTERPRETER || "python";

module.exports = {
  apps : [
    {
      name   : "ap-dag-finalizer",
      script : `${INTERPRETER} ${__dirname}/gunicorn_dag_finalizer_launcher.py`,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      kill_timeout : 3000,
      env: {
        NODE_ENV: NODE_ENV,
        GUNICORN_WORKERS: 20
      }
    },
    {
      name   : "ap-dag-processor",
      script : `${INTERPRETER} ${__dirname}/dag_finalizer.py`,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      kill_timeout : 3000,
      env: {
        NODE_ENV: NODE_ENV
      }
    },
    {
      name   : "ap-backend",
      script : `${INTERPRETER} ${__dirname}/gunicorn_main_launcher.py`,
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
      cwd : `${__dirname}/go-payload-commit-service`,
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
      cwd : `${__dirname}/token-aggregator`,
      max_restarts: MAX_RESTART,
      env: {
        NODE_ENV: NODE_ENV,
      },
      args: "5"
    },
    {
      name   : "ap-proto-indexer",
      script : `${INTERPRETER} ${__dirname}/proto_sliding_window_cacher_service.py`,
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
      cwd : `${__dirname}/dag_verifier`,
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
      cwd : `${__dirname}/go-pruning-archival-service`,
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
