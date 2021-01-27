from copy import deepcopy
import logging
import sys
import json
from config import settings
import aiohttp
from tenacity import retry, wait_exponential, stop_after_attempt, retry_if_exception
import async_timeout
import os

utils_logger = logging.getLogger(__name__)
utils_logger.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(levelname)-8s %(name)-4s %(asctime)s %(msecs)d %(module)s-%(funcName)s: %(message)s")
stream_handler = logging.StreamHandler(sys.stdout)
stream_handler.setFormatter(formatter)
utils_logger.addHandler(stream_handler)


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


async def process_payloads_for_diff(project_id: str, prev_data: dict, cur_data: dict, reader_redis_conn):
    # load diff rules
    r = await reader_redis_conn.get(f'projectID:{project_id}:diffRules')
    diff_rules = None
    if r:
        try:
            diff_rules = json.loads(r)
            print('Got diff rules for project ID')
            print(project_id)
            print(diff_rules)
        except json.JSONDecodeError:
            pass
    if not diff_rules:
        return dict(
            prev_copy=prev_data,
            cur_copy=cur_data,
            payload_changed=dict(),
        )
    key_rules = dict()
    compare_rules = dict()
    for rule in diff_rules:
        if rule['ruleType'] == 'ignore':
            if rule['fieldType'] == 'list':
                key_rules[rule['field']] = {k: rule[k] for k in
                                            ['ignoreMemberFields', 'fieldType', 'ruleType', 'listMemberType']}
            elif rule['fieldType'] == 'map':
                key_rules[rule['field']] = {k: rule[k] for k in ['ignoreMemberFields', 'fieldType', 'ruleType']}
        elif rule['ruleType'] == 'compare':
            if rule['fieldType'] == 'map':
                compare_rules[rule['field']] = {k: rule[k] for k in ['fieldType', 'operation', 'memberFields']}
    prev_copy = clean_map_members(prev_data, key_rules)
    cur_copy = clean_map_members(cur_data, key_rules)
    payload_changed = compare_members(prev_copy, cur_copy, compare_rules)
    return dict(
        prev_copy=prev_copy,
        cur_copy=cur_copy,
        payload_changed=payload_changed
    )


def clean_map_members(data, key_rules):
    data_copy = deepcopy(data)
    for k in data_copy.keys():
        if k and k in key_rules.keys():
            if key_rules[k]['ruleType'] == 'ignore':
                to_be_ignored_fields = key_rules[k]['ignoreMemberFields']
                if key_rules[k]['fieldType'] == 'list':
                    if key_rules[k]['listMemberType'] == 'map':
                        for each_member in data_copy[k]:
                            for del_k in to_be_ignored_fields:
                                if '.' in del_k:
                                    path_trail = del_k.split('.')
                                    p = None  # last but one in path trail
                                    m = each_member.get(path_trail[0])
                                    for idx, sub_k in enumerate(path_trail):
                                        if idx > 0:
                                            # getting the object reference necessary to pop
                                            if not m:
                                                break
                                            p = m
                                            m = m.get(sub_k)
                                    if p:
                                        p.pop(path_trail[-1])
                                else:
                                    each_member.pop(del_k)
                elif key_rules[k]['fieldType'] == 'map':
                    for del_k in to_be_ignored_fields:
                        if '.' in del_k:
                            path_trail = del_k.split('.')
                            p = None  # last but one in path trail
                            m = data_copy[k].get(path_trail[0])
                            # print('Before for: m: ', m)
                            for idx, sub_k in enumerate(path_trail):
                                if idx > 0:
                                    # getting the object reference necessary to pop
                                    if not m:
                                        break
                                    p = m
                                    m = m.get(sub_k)
                            if p:
                                p.pop(path_trail[-1])
                        else:
                            data_copy[k].pop(del_k)
    return data_copy


def compare_members(prev_data, cur_data, compare_rules: dict):
    comparison_results = dict()
    keys_compared_for_rules = set()
    for field_name in compare_rules.keys():
        print('*' * 40)
        print('Applying comparison rule for key')
        print(field_name)
        prev_data_field = prev_data.get(field_name)
        cur_data_field = cur_data.get(field_name)
        print('Comparing between structures...')
        print(prev_data)
        print(cur_data)
        # collect data fields
        prev_data_comparable_values = list()
        cur_data_comparable_values = list()
        # TODO: expand on checks for field type: map, list etc
        if compare_rules[field_name]['fieldType'] == 'map':
            pass
        else:
            continue
        for each_comparable_member in compare_rules[field_name]['memberFields']:
            print('=' * 40 + '\nCollecting value for field')
            print(each_comparable_member)
            if '.' in each_comparable_member:
                path_trail = each_comparable_member.split('.')
                prev_data_member_field = prev_data_field.get(path_trail[0], 0)
                cur_data_member_field = cur_data_field.get(path_trail[0], 0)
                print('Before iterating over path')
                print('prev data member field value')
                print(prev_data_member_field)
                print('cur data member field value')
                print(cur_data_member_field)
                for idx, sub_k in enumerate(path_trail):
                    if idx > 0:
                        prev_data_member_field = prev_data_member_field.get(path_trail[idx], 0)
                        cur_data_member_field = cur_data_member_field.get(path_trail[idx], 0)
                        if idx == len(path_trail) - 1:
                            prev_data_comparable_values.append(prev_data_member_field)
                            cur_data_comparable_values.append(cur_data_member_field)
            else:
                prev_data_comparable_values.append(prev_data_field.get(each_comparable_member, 0))
                cur_data_comparable_values.append(cur_data_field.get(each_comparable_member, 0))
        if compare_rules[field_name]['operation'] == 'add':
            print('Prev data')
            print(prev_data_comparable_values)
            print('Cur data')
            print(cur_data_comparable_values)
            collated_prev_comparable = sum(prev_data_comparable_values)
            collated_cur_comparable = sum(cur_data_comparable_values)
            comparison_results[field_name] = collated_prev_comparable != collated_cur_comparable
        keys_compared_for_rules.add(field_name)
        print('*' * 40)
    if type(prev_data) is dict and type(cur_data) is dict:
        for k in prev_data:
            if k in keys_compared_for_rules:
                continue
            utils_logger.debug('*' * 40)
            utils_logger.debug('Comparing for field without comparison rule | Comparison results')
            utils_logger.debug(k)
            prev_val = prev_data[k]
            cur_val = cur_data.get(k)
            comparison_results[k] = prev_val != cur_val
            utils_logger.debug(comparison_results)
            keys_compared_for_rules.add(k)
            utils_logger.debug('*' * 40)
    return comparison_results


def preprocess_dag(block):
    if 'Height' in block.keys():
        _block = deepcopy(block)
        dag_structure = settings.dag_structure

        dag_structure['height'] = _block.pop('Height')
        dag_structure['prevCid'] = _block.pop('prevCid')
        _block['Data']['cid'] = _block['Data'].pop('Cid')
        _block['Data']['type'] = _block['Data'].pop('Type')
        dag_structure['data'] = _block.pop('Data')
        dag_structure['txHash'] = _block.pop('TxHash')
        dag_structure['timestamp'] = _block.pop('Timestamp')

        return dag_structure
    else:
        return block
