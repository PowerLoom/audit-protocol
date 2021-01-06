from pygate_grpc.client import PowerGateClient
import logging
import sys

deal_logger = logging.getLogger(__name__)
deal_logger.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stream_handler = logging.StreamHandler(sys.stdout)
stream_handler.setFormatter(formatter)
deal_logger.addHandler(stream_handler)
deal_logger.debug("Initialized logger")


def check_job_status(powergate_client: PowerGateClient, staged_cid: str, token: str):
    """
    - Given the cid and token, find out the status of that storage deal
    """

    # Get the cid_info for the cid
    out = powergate_client.data.cid_info([staged_cid], token=token)
    cid_info = out.pop()

    # Check if the deal is still in the executing stage or has its execution completed
    executing_job = cid_info.executingStorageJob
    final_storage_job = cid_info.latestFinalStorageJob

    if executing_job is None:
        if final_storage_job['status'] == "JOB_STATUS_SUCCESS":
            # Mark this storage job as being successfull
            pass
        elif final_storage_job['status'] == "JOB_STATUS_FAILED":
            # Mark this storage job failed
            pass

    elif executing_job['status'] == "JOB_STATUS_EXECUTING":
        # Leave this job as it is and continue onto the next job
        pass

