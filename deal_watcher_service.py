from pygate_grpc.client import PowerGateClient
import logging
import sys
from redis_conn import provide_async_reader_conn_inst, provide_async_writer_conn_inst
from typing import Union
from pydantic import ValidationError
import json
from config import settings
import asyncio
import aioredis
from ipfs_async import client as ipfs_client
import ipfshttpclient
from data_models import ContainerData, SiaSkynetData, SiaRenterData
import coloredlogs

deal_logger = logging.getLogger(__name__)
deal_logger.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stream_handler = logging.StreamHandler(sys.stdout)
stream_handler.setFormatter(formatter)
deal_logger.addHandler(stream_handler)
deal_logger.debug("Initialized logger")

coloredlogs.install(level="DEBUG", logger=deal_logger, stream=sys.stdout)


def check_job_status(powergate_client: PowerGateClient, staged_cid: str, token: str):
    """
    - Given the cid and token, find out the status of that storage deal
    """

    deal_logger.debug("Checking Job Status for cid: ")
    deal_logger.debug(staged_cid)

    # Get the cid_info for the cid
    out = powergate_client.data.cid_info([staged_cid], token=token)
    cid_info = out.pop()

    # Check if the deal is still in the executing stage or has its execution completed
    executing_job = cid_info.executingStorageJob
    final_storage_job = cid_info.latestFinalStorageJob

    if executing_job is None:
        if final_storage_job['status'] == 'JOB_STATUS_FAILED':
            return final_storage_job['status'], final_storage_job['errorCause']
        else:
            return final_storage_job['status'], ""

    else:
        return executing_job['status'], ""


async def process_job(
        container_id: str,
        container_data: ContainerData,
        powergate_client: PowerGateClient,
        writer_redis_conn
):

    # Check the job status for this container from filecoin
    try:
        job_status, status_description = check_job_status(
            powergate_client=powergate_client,
            staged_cid=container_data.backupMetaData.filecoin.stagedCid,
            token=container_data.backupMetaData.filecoin.filecoinToken
        )
    except Exception as eobj:
        logging.error(eobj, exc_info=True)
        return -1, {}

    deal_logger.debug("Retrieved jobStatus and jobStatusDescription from filecoin: ")
    deal_logger.debug(job_status)
    deal_logger.debug(status_description)

    if container_data.backupMetaData.filecoin.jobStatus != job_status:
        # Update the jobStatus and jobStatusDescription of the jobData and store it on redis
        container_data.backupMetaData.filecoin.jobStatus = job_status
        container_data.backupMetaData.filecoin.jobStatusDescription = status_description
        try:
            container_data.convert_to_json()
        except TypeError as terr:
            deal_logger.debug("The container data contains invalid fields.")
            deal_logger.debug(container_data)
            deal_logger.error(terr, exc_info=True)
            return -1, {}

        container_data_key = f"containerData:{container_id}"
        _ = await writer_redis_conn.hmset_dict(
            key=container_data_key,
            **container_data.dict()
        )
        deal_logger.debug("Stored jobStatus and jobDescription on redis")

    return job_status, container_data


def preprocess_container_data(container_data: dict) -> dict:
    """
    - Backup data might contain 'sia' as a backupTarget, which is essentially sia:skynet.
    This change in the naming convention has to be preprocessed
    :return: container_data: dict
                - The updated and preprocessed container_data
    """
    if isinstance(container_data['backupTargets'], str):
        container_data['backupTargets'] = json.loads(container_data['backupTargets'])

    if "sia" in container_data['backupTargets']:

        container_data['backupTargets'].remove("sia")
        container_data['backupTargets'].append("sia:skynet")

        if isinstance(container_data['backupMetaData'], str):
            container_data['backupMetaData'] = json.loads(container_data['backupMetaData'])

        sia_data = container_data['backupMetaData']['sia']
        if isinstance(sia_data, str):
            sia_data = json.loads(sia_data)
        container_data['backupMetaData']['sia_skynet'] = SiaSkynetData(skylink=sia_data['skylink'])

        try:
            del container_data['backupMetaData']['sia']
        except Exception as eobj:
            pass

    return container_data


async def get_container_data(
        container_id: str,
        reader_redis_conn
) -> Union[int, ContainerData]:

    # Get container_data from redis
    container_data_key = f"containerData:{container_id}"
    out = await reader_redis_conn.hgetall(container_data_key)
    if out:
        container_data = {k.decode('utf-8'): v.decode('utf-8') for k, v in out.items()}
    else:
        deal_logger.warning("Data for this container does not exist on redis")
        return -1

    # Preprocess the container_data
    try:
        container_data = preprocess_container_data(container_data=container_data)
    except json.JSONDecodeError as jdecerr:
        deal_logger.warning("Error while converting this container to dict")
        deal_logger.error(jdecerr, exc_info=True)
        return -1

    # Convert the container_data to ContainerData model
    try:
        container_data = ContainerData(**container_data)
    except ValidationError as verr:
        deal_logger.debug("Invalid container_data: ")
        deal_logger.debug(container_data)
        deal_logger.error(verr, exc_info=True)
        return -1

    # Return container_data
    deal_logger.debug(container_data)
    return container_data


async def unpin_cids(
        from_height: int,
        to_height: int,
        project_id: str,
        reader_redis_conn,
):
    """
        - Get the list of payloadCids from redis for the given project_id
        - unpin them one by one
    """

    # Get the list of payloadCid's from redis
    payload_cid_key = f"projectID:{project_id}:payloadCids"
    payload_cids = await reader_redis_conn.zrangebyscore(
        key=payload_cid_key,
        min=from_height,
        max=to_height,
        withscores=False
    )

    # Unpin each payload_cid one by one
    for payload_cid in payload_cids:
        if isinstance(payload_cid, bytes):
            payload_cid = payload_cid.decode('utf-8')
        deal_logger.debug("Unpinning cid: ")
        deal_logger.debug(payload_cid)
        try:
            _ = ipfs_client.pin.rm(payload_cid)
        except ipfshttpclient.exceptions.ErrorResponse as err:
            warning_message = f"Cid: {payload_cid} is not pinned..."
            deal_logger.warning(warning_message)


@provide_async_reader_conn_inst
@provide_async_writer_conn_inst
async def start(
        reader_redis_conn=None,
        writer_redis_conn=None
):

    # Get the list of all projectId's
    failed_containers_key = f"failedContainers"
    executing_containers_key = f"executingContainers"
    all_containers_key = f"containerData:*"
    executing_containers = []

    if settings.UNPIN_MODE == "all":
        out = await reader_redis_conn.keys(all_containers_key)
        if out:
            for container_key in out:
                container_key = container_key.decode('utf-8')
                container_id = container_key.split(':')[-1]
                executing_containers.append(container_id)

    else:
        executing_containers = await reader_redis_conn.smembers(executing_containers_key)

    powergate_client = None
    full_lines = "="*80

    if not executing_containers:
        deal_logger.debug("No executing containers found")
        return 0

    for container_id in executing_containers:
        deal_logger.debug(full_lines)

        if isinstance(container_id, bytes):
            container_id = container_id.decode('utf-8')

        deal_logger.debug("Processing job for container_id: ")
        deal_logger.debug(container_id)

        container_data: Union[int, ContainerData] = await get_container_data(
            container_id=container_id,
            reader_redis_conn=reader_redis_conn
        )

        if container_data == -1:
            deal_logger.debug("Skipping this container..")
            continue

        if "filecoin" in container_data.backupTargets:
            if powergate_client is None:
                powergate_client = PowerGateClient(settings.POWERGATE_CLIENT_ADDR)

            if container_data.backupMetaData.filecoin.jobStatus == "JOB_STATUS_SUCCESS":
                deal_logger.debug("This container has already been processed...")
                continue

            out = await process_job(
                container_id=container_id,
                container_data=container_data,
                powergate_client=powergate_client,
                writer_redis_conn=writer_redis_conn
            )
            job_status: Union[int, str] = out[0]
            container_data: Union[dict, ContainerData] = out[1]

            if job_status == -1:
                deal_logger.warning("Processing job failed")
                continue

            elif job_status == "JOB_STATUS_EXECUTING":
                deal_logger.debug("Job Status Executing....")
                continue

            else:
                _ = await writer_redis_conn.srem(executing_containers_key, container_id)
                deal_logger.debug("Removed the container_id from executingContainers redis SET")

                if job_status == "JOB_STATUS_FAILED":
                    _ = await writer_redis_conn.sadd(failed_containers_key, container_id)
                    deal_logger.debug("Adding the container_id to failedContainers redis SET ")

                if job_status == "JOB_STATUS_SUCCESS":
                    _ = await unpin_cids(
                        from_height=container_data.fromHeight,
                        to_height=container_data.toHeight,
                        project_id=container_data.projectId,
                        reader_redis_conn=reader_redis_conn
                    )
        else:
            deal_logger.debug("Non filecoin container. Backed up to: ")
            deal_logger.debug(container_data.backupTargets)
            deal_logger.debug("Unpinning the cid's")
            _ = await unpin_cids(
                from_height=container_data.fromHeight,
                to_height=container_data.toHeight,
                project_id=container_data.projectId,
                reader_redis_conn=reader_redis_conn
            )

        deal_logger.debug(full_lines)


def crash_done_callback(fut: asyncio.Future):
    try:
        exc = fut.exception()
    except (asyncio.CancelledError, aioredis.ConnectionClosedError):
        deal_logger.debug("Respawning deal watcher")
        t = asyncio.ensure_future(periodic_deal_monitoring())
        t.add_done_callback(crash_done_callback)
    except Exception as eobj:
        deal_logger.debug("There was an error while running the deal watcher: ")
        deal_logger.error(eobj, exc_info=True)


async def periodic_deal_monitoring():
    while True:
        await asyncio.gather(
            start(),
            asyncio.sleep(settings.DEAL_WATCHER_SERVICE_INTERVAL)
        )


if __name__ == "__main__":
    f = asyncio.ensure_future(periodic_deal_monitoring())
    f.add_done_callback(crash_done_callback)
    try:
        asyncio.get_event_loop().run_until_complete(asyncio.gather(f))
    except Exception as e:
        asyncio.get_event_loop().stop()
