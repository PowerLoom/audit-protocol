import aiohttp
import async_timeout
import logging
import hashlib
import os
import json
import time

from pygate_grpc.client import PowerGateClient
import siaskynet
from siaskynet import SkynetClient

from config import settings

from data_models import ContainerData, FilecoinJobData, SiaRenterData, SiaSkynetData
from pydantic import ValidationError
from typing import Union
from tenacity import retry, wait_exponential, stop_after_attempt, retry_if_exception

utils_logger = logging.getLogger(__name__)
utils_logger.setLevel(logging.DEBUG)


class FailedRequestToSiaRenter(Exception):
    """Raised whenever the call to Sia Renter API fails"""
    pass


class FailedRequestToSiaSkynet(Exception):
    """ Raised whenever the call to Sia Skynet Fails."""


@retry(
    wait=wait_exponential(min=2, max=18, multiplier=1),
    stop=stop_after_attempt(6),
    retry=retry_if_exception(FailedRequestToSiaRenter)
)
async def sia_upload(file_hash, file_content):
    headers = {'user-agent': 'Sia-Agent', 'content-type': 'application/octet-stream'}
    utils_logger.debug("Attempting to upload file on Sia...")
    utils_logger.debug(file_hash)
    eobj = None
    async with aiohttp.ClientSession() as session:
        async with async_timeout.timeout(60) as cm:
            try:
                async with session.post(
                        url=f"http://localhost:9980/renter/uploadstream/{file_hash}?datapieces=10&paritypieces=20",
                        headers=headers,
                        data=file_content
                ) as response:
                    utils_logger.debug("Got response from Sia /renter/uploadstream")
                    utils_logger.debug("Response Status: ")
                    utils_logger.debug(response.status)

                    response_text = await response.text()
                    utils_logger.debug("Response Text: ")
                    utils_logger.debug(response_text)
            except Exception as eobj:
                utils_logger.debug("An Exception occurred: ")
                utils_logger.debug(eobj)

            if eobj or cm.expired:
                utils_logger.debug("Retrying post request to /renter/uploadstream")
                raise FailedRequestToSiaRenter("Request to /renter/uploadstream failed")

            if response.status in range(200, 210):
                utils_logger.debug("File content successfully pushed to Sia")
            elif response.status == 500:
                utils_logger.debug("Failed to push the file to Sia")
            else:
                utils_logger.debug("Retrying post request to /renter/uploadstream")
                raise FailedRequestToSiaRenter("Request to /renter/uploadstream failed")


@retry(
    wait=wait_exponential(min=2, max=18, multiplier=1),
    stop=stop_after_attempt(6),
    retry=retry_if_exception(FailedRequestToSiaRenter)
)
async def sia_get(file_hash, force=True):
    """Get the file content for the file hash from Sia"""
    headers = {'user-agent': 'Sia-Agent'}
    file_path = f"temp_files/{file_hash}"
    if (force is True) or (os.path.exists(file_path) is False):
        async with aiohttp.ClientSession() as session:
            async with async_timeout.timeout(6) as cm:
                try:
                    async with session.get(
                            url=f"http://localhost:9980/renter/stream/{file_hash}",
                            headers=headers,
                    ) as response:

                        utils_logger.debug("Got response from Sia /renter/stream")
                        utils_logger.debug("Response status: ")
                        utils_logger.debug(response.status)

                except Exception as eobj:
                    utils_logger.debug("An Exception occured: ")
                    utils_logger.debug(eobj)

                if eobj or cm.expired:
                    raise FailedRequestToSiaRenter("Request to /renter/stream Failed")

                if response.status != 200:
                    raise FailedRequestToSiaRenter("Request to /renter/stream Failed")
                else:
                    utils_logger.debug("File content successfully retrieved from Sia")
                    f = open(file_path, 'ab')
                    async for data in response.content.iter_chunked(n=1024*50):
                        f.write(data)
                    f.close()
    f = open(file_path, 'rb')
    data = f.read()
    return data.decode('utf-8')


async def get_data_from_filecoin(filecoin_job_data: FilecoinJobData):
    powgate_client = PowerGateClient(settings.powergate_client_addr, False)
    out = powgate_client.data.get(filecoin_job_data.stagedCid, token=filecoin_job_data.filecoinToken).decode('utf-8')
    container = json.loads(out)['container']
    return container


@retry(
    wait=wait_exponential(min=2, max=18, multiplier=1),
    stop=stop_after_attempt(6),
    retry=retry_if_exception(FailedRequestToSiaSkynet)
)
async def get_data_from_sia_skynet(sia_data: SiaSkynetData, container_id: str):
    utils_logger.debug("Getting container from Sia")
    utils_logger.debug(sia_data.skylink)
    client = SkynetClient()
    timestamp = int(time.time())
    temp_path = f"temp_files/{container_id}"
    if not os.path.exists(temp_path):
        try:
            client.download_file(skylink=sia_data.skylink,  path=temp_path)
        except Exception as e:
            utils_logger.debug("Failed to get data from Sia Skynet")
            raise FailedRequestToSiaSkynet

    f = open(temp_path, 'r')
    data = f.read()
    try:
        json_data = json.loads(data)
    except json.JSONDecodeError as jdecerr:
        utils_logger.debug("An error occured while loading data from skynet")
        utils_logger.error(jdecerr, exc_info=True)
        return -1
    return json_data['container']


@retry(
    wait=wait_exponential(min=2, max=18, multiplier=1),
    stop=stop_after_attempt(6),
    retry=retry_if_exception(FailedRequestToSiaRenter)
)
async def get_data_from_sia_renter(sia_renter_data: SiaRenterData, container_id: str):
    try:
        out = await sia_get(sia_renter_data.fileHash)
    except FailedRequestToSiaRenter as ferr:
        utils_logger.debug("Retrying to get the data from sia renter")
        raise FailedRequestToSiaRenter

    try:
        container_data = json.loads(out)
    except json.JSONDecodeError as jdecerr:
        utils_logger.debug("There was an error while loading json data.")
        utils_logger.error(jdecerr, exc_info=True)
        return -1

    return container_data['container']


async def get_backup_data(container_data: dict, container_id: str):
    data = None
    backupTargets = []
    if isinstance(container_data['backupTargets'], str):
        backupTargets = json.loads(container_data['backupTargets'])
    if isinstance(container_data['backupMetaData'], str):
        container_data['backupMetaData'] = json.loads(container_data['backupMetaData'])
    if "sia" in backupTargets:
        backupTargets.remove("sia")
        backupTargets.append("sia:skynet")

        sia_data = container_data['backupMetaData']['sia']
        if isinstance(sia_data, str):
            sia_data = json.loads(sia_data)
        container_data['backupTargets'] = backupTargets
        container_data['backupMetaData']['sia_skynet'] = SiaSkynetData(skylink=sia_data['skylink'])
        try:
            del container_data['backupMetaData']['sia']
        except Exception as e:
            pass

    try:
        container_data = ContainerData(**container_data)
    except ValidationError as verr:
        utils_logger.debug("There was an error while trying to create ContainerData model")
        utils_logger.error(verr, exc_info=True)
        return -1

    if "filecoin" in container_data.backupTargets:
        data = await get_data_from_filecoin(container_data.backupMetaData.filecoin)
    elif "sia:skynet" in container_data.backupTargets:
        data = await get_data_from_sia_skynet(container_data.backupMetaData.sia_skynet, container_id=container_id)
    elif "sia:renter" in container_data.backupTargets:
        data = await get_data_from_sia_renter(container_data.backupMetaData.sia_renter, container_id=container_id)

    return data


async def backup_to_filecoin(
        container_data: dict,
        reader_redis_conn=None,
        writer_redis_conn=None
) -> Union[int, FilecoinJobData]:
    """
        - Convert container data to json string.
        - Push it to filecoin
        - Save the FilecoinJobData
    """
    powgate_client = PowerGateClient(settings.powergate_client_addr, False)

    """ Get the token """
    filecoin_token_key = f"filecoinToken:{container_data['projectId']}"
    token = await reader_redis_conn.get(filecoin_token_key)
    if not token:
        user = powgate_client.admin.users.create()
        token = user.token
        _ = await writer_redis_conn.set(filecoin_token_key, token)

    else:
        token = token.decode('utf-8')

    """ Convert the data to json string """
    try:
        json_data = json.dumps(container_data).encode('utf-8')
    except TypeError as terr:
        utils_logger.debug("Unable to convert container data to json string")
        utils_logger.error(terr, exc_info=True)
        return -1

    # Stage and push the data
    try:
        stage_res = powgate_client.data.stage_bytes(json_data, token=token)
        job = powgate_client.config.apply(stage_res.cid, override=True, token=token)
    except Exception as eobj:
        utils_logger.debug("Failed to backup container data to filecoin")
        utils_logger.error(eobj, exc_info=True)
        return -1

    # Save the jobId and other data
    job_data = {
        'stagedCid': stage_res.cid,
        'jobId': job.jobId,
        'jobStatus': 'JOB_STATUS_UNCHECKED',
        'jobStatusDescription': "",
        'filecoinToken': token,
        'retries': 0,
    }
    try:
        filecoin_job_data = FilecoinJobData(**job_data)
        return filecoin_job_data
    except ValidationError as verr:
        utils_logger.debug("Invalid data passed to FilecoinJobData")
        utils_logger.error(verr, exc_info=True)
    except json.JSONDecodeError as jerr:
        utils_logger.debug("Invalid Json Data passed to FilecoinJobData")
        utils_logger.error(jerr, exc_info=True)

    return -1


async def backup_to_sia_renter(container_data: dict):
    """
    - Backup the given container data to Sia Renter
    """

    # Convert container data to Json
    utils_logger.debug("Backing up data to Sia Renter")
    try:
        json_data = json.dumps(container_data)
    except TypeError as terr:
        utils_logger.debug("Failed to convert container data to json string")
        utils_logger.error(terr, exc_info=True)
        return -1

    file_hash = hashlib.md5(json_data.encode('utf-8')).hexdigest()
    try:
        _ = await sia_upload(file_hash=file_hash, file_content=json_data.encode('utf-8'))
    except FailedRequestToSiaRenter as ferr:
        utils_logger.debug("Failed to push data to Sia Renter")
        utils_logger.debug(ferr, exc_info=True)
        return -1

    try:
        sia_renter_ata = SiaRenterData(fileHash=file_hash)
    except ValidationError as verr:
        utils_logger.debug("Failed to convert sia renter data into a model")
        utils_logger.error(verr, exc_info=True)
        return -1

    return sia_renter_ata


async def backup_to_sia_skynet(container_data: dict):
    """
    - Backup the given container data onto Sia
    """

    utils_logger.debug("Backing up data to Sia Skynet")
    container_id = container_data['containerId']
    container_file_path = f"containers/{container_id}.json"
    client = siaskynet.SkynetClient()
    try:
        skylink = client.upload_file(container_file_path)
    except Exception as e:
        utils_logger.debug("There was an error while uploading the file to Skynet")
        utils_logger.error(e, exc_info=True)
        return -1

    try:
        sia_skynet_data = SiaSkynetData(skylink=skylink)
    except ValidationError as verr:
        utils_logger.debug("Failed to convert sia skynet data into a model")
        utils_logger.error(verr, exc_info=True)
        return -1

    return sia_skynet_data
