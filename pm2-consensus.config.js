// this means if app restart {MAX_RESTART} times in 1 min then it stops
const NODE_ENV = process.env.NODE_ENV || 'development';
const INTERPRETER = process.env.CONSENSUS_INTERPRETER || "python";
const BEGIN_BLOCK = process.env.TICKER_BEGIN_BLOCK || "";


const MAX_RESTART = 10;
const MIN_UPTIME = 60000;


module.exports = {
  apps : [
    {
      name   : "system-ticker-linear",
      script : `${INTERPRETER} -m snapshot_consensus.system_ticker_linear ${BEGIN_BLOCK } `,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      env: {
        NODE_ENV: NODE_ENV,
      }
    },
    {
      name   : "offchain-consensus-api",
      script : `uvicorn snapshot_consensus.centralized:app --reload`,
      max_restarts: MAX_RESTART,
      min_uptime: MIN_UPTIME,
      env: {
        NODE_ENV: NODE_ENV,
      }
    }
  ]
}
