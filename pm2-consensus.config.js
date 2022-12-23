// this means if app restart {MAX_RESTART} times in 1 min then it stops

const { readFileSync } = require('fs');
const settings = JSON.parse(readFileSync('snapshot_consensus/settings.json'));

const NODE_ENV = process.env.NODE_ENV || 'development';
const INTERPRETER = process.env.CONSENSUS_INTERPRETER || "python";
const BEGIN_BLOCK = process.env.TICKER_BEGIN_BLOCK || "";


const MAX_RESTART = 10;
const MIN_UPTIME = 60000;


module.exports = {
  apps : [
    {
      name   : "epoch-tracker",
      script : `${INTERPRETER} -m snapshot_consensus.epoch_generator ${BEGIN_BLOCK } `,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      env: {
        NODE_ENV: NODE_ENV,
      }
    },
    {
      name   : "off-chain-consensus",
      script: `gunicorn -k uvicorn.workers.UvicornWorker snapshot_consensus.centralized:app --workers 4 -b ${settings.consensus_service.host}:${settings.consensus_service.port}`,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      env: {
        NODE_ENV: NODE_ENV,
      }
    }
  ]
}
