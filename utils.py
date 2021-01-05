from copy import deepcopy
from logging import getLogger
import json


logger = getLogger(__name__)


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
                key_rules[rule['field']] = {k: rule[k] for k in ['ignoreMemberFields', 'fieldType', 'ruleType', 'listMemberType']}
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
        print('*'*40)
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
            print('='*40+'\nCollecting value for field')
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
