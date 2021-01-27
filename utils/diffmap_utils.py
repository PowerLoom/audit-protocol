import json
from copy import deepcopy
import logging
from dynaconf import settings

from utils import dag_utils
from utils import redis_keys
from utils.redis_conn import provide_async_reader_conn_inst, provide_async_writer_conn_inst

logger = logging.getLogger(__name__)
logger.setLevel(level="DEBUG")


@provide_async_reader_conn_inst
async def get_diff_rules(
        project_id: str,
        reader_redis_conn
) -> list:
    diff_rules_key = redis_keys.get_diff_rules_key(project_id=project_id)
    out: bytes = await reader_redis_conn.get(diff_rules_key)
    diff_rules = []

    if out:
        diff_rules = out.decode('utf-8')
        try:
            diff_rules = json.loads(diff_rules_key)
        except json.JSONDecodeError as jerr:
            logger.error(jerr, exc_info=True)

    return diff_rules


async def process_payloads_for_diff(project_id: str, prev_data: dict, cur_data: dict):
    # load diff rules
    diff_rules = await get_diff_rules(project_id)
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
    payload_changed = dict()
    if len(compare_rules) > 0:
        payload_changed = compare_members(prev_copy, cur_copy, compare_rules)
    # else:
    #     for k, v in cur_copy.items():
    #         payload_changed[k] = (prev_copy.get(k) == v)
    #
    #     print(payload_changed)

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


def compare_members(prev_data, cur_data, compare_rules):
    comparison_results = dict()
    for field_name in compare_rules:
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
        # TODO: check for field type: map, list etc
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
        print('*' * 40)
    return comparison_results


@provide_async_writer_conn_inst
async def calculate_diff(
        dag_cid: str,
        dag: dict,
        project_id: str,
        ipfs_client,
        writer_redis_conn
):
    # cache last seen diffs
    dag_height = dag['height']
    if dag['prevCid']:
        payload_cid = dag['data']['cid']
        prev_dag = await dag_utils.get_dag_block(dag['prevCid'])
        prev_payload_cid = prev_dag['data']['cid']
        if prev_payload_cid != payload_cid:
            diff_map = dict()
            _prev_data = await ipfs_client.cat(prev_payload_cid)
            prev_data = _prev_data.decode('utf-8')
            prev_data = json.loads(prev_data)
            _payload = await ipfs_client.cat(payload_cid)
            payload = _payload.decode('utf-8')
            payload = json.loads(payload)
            logger.debug('Before payload clean up')
            logger.debug({'cur_payload': payload, 'prev_payload': prev_data})
            result = await process_payloads_for_diff(
                project_id,
                prev_data,
                payload,
            )
            logger.debug('After payload clean up and comparison if any')
            logger.debug(result)
            cur_data_copy = result['cur_copy']
            prev_data_copy = result['prev_copy']

            for k, v in cur_data_copy.items():
                if k in result['payload_changed'] and result['payload_changed'][k]:
                    diff_map[k] = {'old': prev_data.get(k), 'new': payload.get(k)}

            if diff_map:
                # rest_logger.debug("DAG at point B:")
                # rest_logger.debug(dag)
                diff_data = {
                    'cur': {
                        'height': dag_height,
                        'payloadCid': payload_cid,
                        'dagCid': dag_cid,
                        'txHash': dag['txHash'],
                        'timestamp': dag['timestamp']
                    },
                    'prev': {
                        'height': prev_dag['height'],
                        'payloadCid': prev_payload_cid,
                        'dagCid': dag['prevCid']
                        # this will be used to fetch the previous block timestamp from the DAG
                    },
                    'diff': diff_map
                }
                if settings.METADATA_CACHE == 'redis':
                    diff_snapshots_cache_zset = f'projectID:{project_id}:diffSnapshots'
                    await writer_redis_conn.zadd(
                        diff_snapshots_cache_zset,
                        score=int(dag['height']),
                        member=json.dumps(diff_data)
                    )
                    latest_seen_snapshots_htable = 'auditprotocol:lastSeenSnapshots'
                    await writer_redis_conn.hset(
                        latest_seen_snapshots_htable,
                        project_id,
                        json.dumps(diff_data)
                    )

                return diff_data

    return {}


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
