from utils import dag_utils
from utils import redis_keys
from config import settings
from copy import deepcopy
import aioredis
import json
import logging.handlers
import sys

utils_logger = logging.getLogger(__name__)
utils_logger.setLevel(level="DEBUG")
formatter = logging.Formatter(u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")

stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
stdout_handler.setFormatter(formatter)

stderr_handler = logging.StreamHandler(sys.stderr)
stderr_handler.setLevel(logging.ERROR)
stderr_handler.setFormatter(formatter)

utils_logger.addHandler(stdout_handler)
utils_logger.addHandler(stderr_handler)


async def get_diff_rules(
        project_id: str,
        reader_redis_conn
) -> list:
    diff_rules_key = redis_keys.get_diff_rules_key(project_id=project_id)
    out: bytes = await reader_redis_conn.get(diff_rules_key)
    diff_rules = list()

    if out:
        diff_rules = out.decode('utf-8')
        try:
            diff_rules = json.loads(diff_rules)
        except json.JSONDecodeError as jerr:
            utils_logger.error(jerr, exc_info=True)
            diff_rules = list()
    return diff_rules


async def process_payloads_for_diff(project_id: str, prev_data: dict, cur_data: dict, redis_conn: aioredis.Redis):
    # load diff rules
    diff_rules = await get_diff_rules(project_id, redis_conn)
    key_rules = dict()
    compare_rules = dict()
    # TODO: refactor logic into another helper that solely preprocesses key rules into an accepted structure
    for rule in diff_rules:
        if rule['ruleType'] == 'ignore':
            if rule['fieldType'] == 'list':
                key_rules[rule['field']] = {k: rule[k] for k in
                                            ['ignoreMemberFields', 'fieldType', 'ruleType', 'listMemberType']}
            elif rule['fieldType'] == 'map':
                key_rules[rule['field']] = {k: rule[k] for k in ['ignoreMemberFields', 'fieldType', 'ruleType']}
            else:
                key_rules[rule['field']] = {k: rule[k] for k in ['fieldType', 'ruleType']}

        elif rule['ruleType'] == 'compare':
            if rule['fieldType'] == 'map':
                compare_rules[rule['field']] = {k: rule[k] for k in ['fieldType', 'operation', 'memberFields']}
        else:
            key_rules[rule['field']] = {'fieldType': rule['fieldType']}
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
    top_level_keys_to_be_deleted = set()
    for k in data_copy.keys():
        if k and k in key_rules.keys():
            # print(f'Processing key rule for {k}')
            if key_rules[k]['ruleType'] == 'ignore':
                # TODO: add support for elementary data types. Support typing module for standardized config?
                if key_rules[k]['fieldType'] in ['str', 'int', 'float']:
                    # collect to be deleted later, we can not delete keys while iterating over the dict
                    top_level_keys_to_be_deleted.add(k)
                    # print(f'Adding key {k} in set of top level keys to be deleted from data_copy')
                    continue
                to_be_ignored_fields = key_rules[k]['ignoreMemberFields']
                if key_rules[k]['fieldType'] == 'list':
                    if key_rules[k]['listMemberType'] == 'map' and k in data_copy.keys():
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
                                        try:
                                            p.pop(path_trail[-1])
                                        except:
                                            pass
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
                                try:
                                    p.pop(path_trail[-1])
                                except:
                                    pass
                        else:
                            try:
                                data_copy[k].pop(del_k)
                            except :
                                pass
    for tbd_k in top_level_keys_to_be_deleted:
        # print(f'Deleting key {tbd_k} from data_copy')
        del data_copy[tbd_k]
        # print(f'Cur data copy: {data_copy}')
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
        collect_comparable_values = True
        # TODO: expand on checks for field type: map, list etc
        if compare_rules[field_name]['fieldType'] == 'map':
            pass
        else:
            continue
        if compare_rules[field_name]['operation'] == 'listSlice':
            collect_comparable_values = False
        if collect_comparable_values:
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
        else:
            map_keys_arr_index = compare_rules[field_name]['memberFields']
            prev_data_comparable_value = prev_data_field.get(list(prev_data_field.keys())[map_keys_arr_index])
            cur_data_comparable_value = cur_data_field.get(list(cur_data_field.keys())[map_keys_arr_index])
            print('Prev data')
            print(prev_data_comparable_value)
            print('Cur data')
            print(cur_data_comparable_value)
            comparison_results[field_name] = prev_data_comparable_value != cur_data_comparable_value
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


async def calculate_diff(
        dag_cid: str,
        dag: dict,
        project_id: str,
        ipfs_client,
        writer_redis_conn: aioredis.Redis
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
            utils_logger.debug('Before payload clean up')
            utils_logger.debug({'cur_payload': payload, 'prev_payload': prev_data})
            result = await process_payloads_for_diff(
                project_id,
                prev_data,
                payload,
                writer_redis_conn
            )
            utils_logger.debug('After payload clean up and comparison if any')
            utils_logger.debug(result)
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
                diff_snapshots_cache_zset = f'projectID:{project_id}:diffSnapshots'
                await writer_redis_conn.zadd(
                    name=diff_snapshots_cache_zset,
                    mapping={json.dumps(diff_data): int(dag['height'])}
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
    # legacy clean up for DAG blocks with an older data structure, most likely not necessary
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
