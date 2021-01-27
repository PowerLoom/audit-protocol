import json
import requests
from urllib.parse import urljoin
import time

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


host_url = "http://localhost:9000/"
diff_rules_url = '/{}/diffRules'
commit_payload_url = '/commit_payload'


def set_diff_rules(project_id: str):

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


set_diff_rules("SomeProject")
commit_payload("SomeProject")