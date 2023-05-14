// this means if app restart {MAX_RESTART} times in 1 min then it stops
const MAX_RESTART = 10;
const MIN_UPTIME = 60000;
const NODE_ENV = process.env.NODE_ENV || 'development';

module.exports = {
  apps : [
    {
      name   : "ap-payload-commit",
      script : "./payload-commit",
      cwd : `${__dirname}/go/payload-commit`,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      kill_timeout : 3000,
      env: {
        NODE_ENV: NODE_ENV,
        CONFIG_PATH:`${__dirname}`,
        PRIVATE_KEY: process.env.PRIVATE_KEY
      },
      args: "5" //Log level set to debug, for production change to 4 (INFO) or 2(ERROR)
    },
    {
      name   : "ap-pruning",
      script : "./pruning",
      cwd : `${__dirname}/go/pruning`,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      kill_timeout : 3000,
      env: {
        NODE_ENV: NODE_ENV,
        CONFIG_PATH:`${__dirname}`,
      },
      args: "5" //Log level set to debug, for production change to 4 (INFO) or 2(ERROR)
    }
  ]
}
