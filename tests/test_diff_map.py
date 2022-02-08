from utils.diffmap_utils import clean_map_members, process_payloads_for_diff
import asyncio
import requests
from urllib.parse import urljoin
import time
import os


rules = {
    'rules': [
        {
            'ruleType': 'compare',
            'field': 'test_field',
            'fieldType': 'map',
            'operation': 'add',
            'memberFields': ['a.b', 'c.d']
        },
    ]
}

FPMM_POOLER_RULES = {
    'rules': [
        {
            "ruleType": "ignore",
            "field": "chainHeightRange",
            "fieldType": "map",
            "ignoreMemberFields": ["begin", "end"],
        },
        {
            "ruleType": "compare",
            "field": "token0Reserves",
            "fieldType": "map",
            "operation": "listSlice",
            "memberFields": -1
        },
        {
            "ruleType": "compare",
            "field": "token1Reserves",
            "fieldType": "map",
            "operation": "listSlice",
            "memberFields": -1
        },

        {
            "ruleType": "ignore",
            "field": "broadcast_id",
            "fieldType": "str"
        },

        {
            "ruleType": "ignore",
            "field": "timestamp",
            "fieldType": "float"
        }
    ]
}

PAYLOAD_OLD = {
        "contract": "0x853ee4b2a13f8a742d64c8f088be7ba2131f670d",
        "token0Reserves": {
            "block24540280": 24349146.952458,
            "block24540281": 24349146.952458,
            "block24540282": 24349146.952458,
            "block24540283": 24349146.952458,
            "block24540284": 24349146.952458,
            "block24540285": 24349146.952458,
            "block24540286": 24349146.952458,
            "block24540287": 24349146.952458,
            "block24540288": 24349146.952458,
            "block24540289": 24349146.952458,
            "block24540290": 24349146.952458,
            "block24540291": 24349146.952458,
            "block24540292": 24349146.952458,
            "block24540293": 24349146.952458,
            "block24540294": 24349146.952458,
            "block24540295": 24349146.952458,
            "block24540296": 24349146.952458,
            "block24540297": 24349146.952458,
            "block24540298": 24349146.952458,
            "block24540299": 24349146.952458,
            "block24540300": 24349146.952458
        },
        "token1Reserves": {
            "block24540280": 8584.903022528188,
            "block24540281": 8584.903022528188,
            "block24540282": 8584.903022528188,
            "block24540283": 8584.903022528188,
            "block24540284": 8584.903022528188,
            "block24540285": 8584.903022528188,
            "block24540286": 8584.903022528188,
            "block24540287": 8584.903022528188,
            "block24540288": 8584.903022528188,
            "block24540289": 8584.903022528188,
            "block24540290": 8584.903022528188,
            "block24540291": 8584.903022528188,
            "block24540292": 8584.903022528188,
            "block24540293": 8584.903022528188,
            "block24540294": 8584.903022528188,
            "block24540295": 8584.903022528188,
            "block24540296": 8584.903022528188,
            "block24540297": 8584.903022528188,
            "block24540298": 8584.903022528188,
            "block24540299": 8584.903022528188,
            "block24540300": 8584.903022528188
        },
        "chainHeightRange": {
            "begin": 24540280,
            "end": 24540300
        },
        "broadcast_id": "sdfsdfsdf",
        "timestamp": 1643977765.6607
    }

PAYLOAD_NEW = {
    "contract": "0x853ee4b2a13f8a742d64c8f088be7ba2131f670d",
    "token0Reserves": {
        "block24540280": 24349146.952458,
        "block24540281": 24349146.952458,
        "block24540282": 24349146.952458,
        "block24540283": 24349146.952458,
        "block24540284": 24349146.952458,
        "block24540285": 24349146.952458,
        "block24540286": 24349146.952458,
        "block24540287": 24349146.952458,
        "block24540288": 24349146.952458,
        "block24540289": 24349146.952458,
        "block24540290": 24349146.952458,
        "block24540291": 24349146.952458,
        "block24540292": 24349146.952458,
        "block24540293": 24349146.952458,
        "block24540294": 24349146.952458,
        "block24540295": 24349146.952458,
        "block24540296": 24349146.952458,
        "block24540297": 24349146.952458,
        "block24540298": 24349146.952458,
        "block24540299": 24349146.952458,
        "block24540300": 24049146.952458
    },
    "token1Reserves": {
        "block24540280": 8584.903022528188,
        "block24540281": 8584.903022528188,
        "block24540282": 8584.903022528188,
        "block24540283": 8584.903022528188,
        "block24540284": 8584.903022528188,
        "block24540285": 8584.903022528188,
        "block24540286": 8584.903022528188,
        "block24540287": 8584.903022528188,
        "block24540288": 8584.903022528188,
        "block24540289": 8584.903022528188,
        "block24540290": 8584.903022528188,
        "block24540291": 8584.903022528188,
        "block24540292": 8584.903022528188,
        "block24540293": 8584.903022528188,
        "block24540294": 8584.903022528188,
        "block24540295": 8584.903022528188,
        "block24540296": 8584.903022528188,
        "block24540297": 8584.903022528188,
        "block24540298": 8584.903022528188,
        "block24540299": 8584.903022528188,
        "block24540300": 8504.903022528188
    },
    "chainHeightRange": {
        "begin": 24540301,
        "end": 24540320
    },
    "broadcast_id": "sdvx4x",
    "timestamp": 1643977775.6607
}

external_ip = os.getenv("EXTERNAL_IP", "localhost")
host_url = f"http://{external_ip}:9000/"
diff_rules_url = '/{}/diffRules'
commit_payload_url = '/commit_payload'


def set_diff_rules(project_id: str, rules: dict):
    target_url = urljoin(host_url, diff_rules_url).format(project_id)

    out = requests.post(url=target_url, json=rules)

    print(out.status_code)
    print(out.text)


def commit_payload(project_id: str):
    payload = {
        'constant_field': 'FIELD_A',
        'test_field': {
            'a': {'b': 20},
            'c': {'d': 30}
        }
    }
    data = {
        'projectId': project_id,
        'payload': payload
    }

    target_url = urljoin(host_url, commit_payload_url)
    for i in range(2):
        data['payload']['test_field']['a']['b'] = i

    out = requests.post(url=target_url, json=data)
    print(out.text)
    time.sleep(30)


def test_pooler_data_cleanup():
    key_rules = dict()
    # processing rule structure
    for rule in FPMM_POOLER_RULES['rules']:
        if rule['ruleType'] == 'ignore':
            if rule['fieldType'] == 'list':
                key_rules[rule['field']] = {k: rule[k] for k in
                                            ['ignoreMemberFields', 'fieldType', 'ruleType', 'listMemberType']}
            elif rule['fieldType'] == 'map':
                key_rules[rule['field']] = {k: rule[k] for k in ['ignoreMemberFields', 'fieldType', 'ruleType']}
            else:
                key_rules[rule['field']] = {k: rule[k] for k in ['fieldType', 'ruleType']}
    r = clean_map_members(data=PAYLOAD_OLD, key_rules=key_rules)
    # print(r)
    assert 'timestamp' not in r.keys() and \
        'broadcast_id' not in r.keys() and \
        'chainHeightRange' in r.keys() and len(r['chainHeightRange']) == 0


def test_pooler_diff():
    set_diff_rules('UniswapPoolerTestProject', FPMM_POOLER_RULES)
    f = process_payloads_for_diff(
        project_id='UniswapPoolerTestProject',
        prev_data=PAYLOAD_OLD,
        cur_data=PAYLOAD_NEW
    )
    r, = asyncio.get_event_loop().run_until_complete(asyncio.gather(f))
    print(r)
    assert r.get('payload_changed') and \
           'token0Reserves' in r['payload_changed'] and r['payload_changed']['token0Reserves'] and \
           'token1Reserves' in r['payload_changed'] and r['payload_changed']['token1Reserves']


def start_commit_diff_test():
    set_diff_rules("SomeProject")
    commit_payload("SomeProject")
