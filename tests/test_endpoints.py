import requests
import time
from test_logger import test_logger
import random
import os

EXTERNAL_IP = os.getenv("EXTERNAL_IP", "localhost")

def test_commit_payloads():
    project_id = "PROJECT_COMMIT_PAYLOAD"
    test_logger.debug("Using the project Id: ")
    test_logger.debug(project_id)
    for i in range(4):
        for j in range(3):
            test_logger.debug("Using i and j: {}, {}, height: {}".format(i, j, (i * 5) + (j + 1)))
            data = {'data': 'data_' + str(i + 1), 'constant_string': "CONSTANT_STRING_HOLDER"}
            payload = {'payload': data, 'projectId': project_id}
            test_logger.debug("Committing payload: ")
            test_logger.debug(payload)
            out = requests.post(f'http://{EXTERNAL_IP}:9000/commit_payload', json=payload)
            assert out.status_code == 200, f"Received {out.status_code} from /commit_payload endpoint"
            data = out.json()
            test_logger.debug("Got Response: ")
            test_logger.debug(data)
            keys = list(data.keys())
            for k in ['cid', 'tentativeHeight', 'payloadChanged']:
                assert k in keys, "Invalid response received from the /commit_payload endpoint"

            assert data['tentativeHeight'] == (i * 3) + (j + 1), \
                "Invalid tentativeHeight from /commit_payload endpoint response"

            if j == 0:
                assert data['payloadChanged'] is True, \
                    "Invalid payload_changed flag received from /commit_payload endpoint"
            else:
                assert data['payloadChanged'] is False, \
                    "Invalid payload_changed flag received from /commit_payload endpoint"


def test_diff_maps():
    project_id = "PROJECT_TEST_DIFF_MAPS"
    test_logger.debug("Using the project Id: ")
    test_logger.debug(project_id)
    for i in range(5):
        for j in range(2):
            old_height = requests.get(f'http://{EXTERNAL_IP}:9000/{project_id}/payloads/height')
            old_height = old_height.json()['height']
            test_logger.debug("Using i and j: {}, {}, old_height: {}".format(i, j, old_height))
            data = {'data': 'data_' + str(i + 1), 'constant_string': "CONSTANT_STRING_HOLDER"}
            payload = {'payload': data, 'projectId': project_id}
            test_logger.debug("Committing payload: ")
            test_logger.debug(payload)
            out = requests.post(f'http://{EXTERNAL_IP}:9000/commit_payload', json=payload)
            assert out.status_code == 200, f"Received {out.status_code} from /commit_payload endpoint"
            data = out.json()
            test_logger.debug("Got Response: ")
            test_logger.debug(data)
            while True:
                new_height = requests.get(f'http://{EXTERNAL_IP}:9000/{project_id}/payloads/height')
                new_height = new_height.json()['height']
                if new_height == old_height:
                    test_logger.debug("Chain has not been updated: {}, {}".format(old_height, new_height))
                    time.sleep(3)
                else:
                    test_logger.debug("Chain has been updated: {}, {}".format(old_height, new_height))
                    break
            if (j == 0) and (i != 0):
                test_logger.debug("Getting the cachedDiff count from the audit-protocol")
                out = requests.get(f"http://{EXTERNAL_IP}:9000/{project_id}/payloads/cachedDiffs/count")
                out = out.json()
                test_logger.debug("Got response:")
                test_logger.debug(out)

                count = out['count']
                assert count == i, "Error! Getting invalid count from cachedDiffs"
                test_logger.debug("cachedDiffs Count Test passed successfully")


def get_diff_rules():
    test_rules = [
        {
            'ruleType': 'ignore',
            'field': 'field_a',
            'fieldType': 'list',
            'listMemberType': 'map',
            'ignoreMemberFields': ['field_b', 'field_c', 'field_d', 'field_f'],
            'payload_changes': False
        },

        {
            'ruleType': 'ignore',
            'field': 'field_a',
            'fieldType': 'list',
            'listMemberType': 'map',
            'ignoreMemberFields': ['field_b', 'field_c.field_d.field_e', 'field_c.field_d.field_f', 'field_f'],
            'payload_changes': False

        }
    ]

    rule = random.choice(test_rules)
    payload_changes = rule.pop('payload_changes')
    return [rule], payload_changes


def set_diff_rules(project_id: str, diff_rules: list):
    test_logger.debug(f"Setting diff rules for {project_id}")
    test_logger.debug(diff_rules)
    url_endpoint = f"http://{EXTERNAL_IP}:9000/{project_id}/diffRules"
    diff_rules = {'rules': diff_rules}
    out = requests.post(url_endpoint, json=diff_rules)
    test_logger.debug("Got response: ")
    test_logger.debug(out.text)
    assert out.status_code == 200, "Request to set rules was unsuccessful"


def test_diff_rules():
    """
    Goals of this test function:
        - Check if the diff count endpoint is working properly
        - Set different diff Rules and then check if they are properly ignoring the required fields
    """

    test_logger.debug("TESTING DIFF RULES")

    project_id = "PROJECT_TEST_DIFF_MAPS"
    test_logger.debug("Using the project Id: ")
    test_logger.debug(project_id)
    total_diff_count = 0
    for i in range(3):
        for j in range(2):
            test_logger.debug("\n")
            test_logger.debug("="*80)
            """ Get the height of the projectId"""
            old_height = requests.get(f'http://{EXTERNAL_IP}:9000/{project_id}/payloads/height')
            old_height = old_height.json()['height']
            test_logger.debug("Using i and j: {}, {}, old_height: {}".format(i, j, old_height))

            """ The payload will have a field that will changed with every new payload and a constant payload """
            data = {
                "constant_string": "CONSTANT_STRING_HOLDER",
                "field_a": [
                    {
                        'field_b': 'data_' + str(i + 1),
                        'field_c': {
                            'field_d': {
                                'field_e': str(i + 1),
                                'field_f': str(i + 1),
                                'field_g': "CONSTANT STRING"
                            }
                        },
                        'field_d': "CONSTANT",
                        'field_f': str(i + 1)
                    }
                ]
            }

            """ Set the diff-rules at a particular point """
            payload_changes = True
            if (i == 2) and (j == 0):
                diff_rules, payload_changes = get_diff_rules()
                set_diff_rules(project_id=project_id, diff_rules=diff_rules)

            payload = {'payload': data, 'projectId': project_id}
            test_logger.debug("Committing payload: ")
            test_logger.debug(payload)

            """ Commit the payload """
            out = requests.post(f'http://{EXTERNAL_IP}:9000/commit_payload', json=payload)
            assert out.status_code == 200, f"Received {out.status_code} from /commit_payload endpoint"
            data = out.json()
            test_logger.debug("Got Response: ")
            test_logger.debug(data)

            """ Wait till the transaction goes through and a new DAG block is created """
            while True:
                new_height = requests.get(f'http://{EXTERNAL_IP}:9000/{project_id}/payloads/height')
                new_height = new_height.json()['height']
                if new_height == old_height:
                    test_logger.debug("Chain has not been updated: {}, {}".format(old_height, new_height))
                    time.sleep(3)
                else:
                    test_logger.debug("Chain has been updated: {}, {}".format(old_height, new_height))
                    break

            if (j == 0) and (i != 0):  # This condition is to check if this is not the first payload
                test_logger.debug("Getting the cachedDiff count from the audit-protocol")
                out = requests.get(f"http://{EXTERNAL_IP}:9000/{project_id}/payloads/cachedDiffs/count")
                out = out.json()
                test_logger.debug("Got response:")
                test_logger.debug(out)

                """ Get the count of the diff maps for this projectId """
                count = out['count']
                if i == 2:
                    count = out['count']
                    if payload_changes is False:
                        assert count == total_diff_count, "Error Diff rules not working"
                    else:
                        assert count == total_diff_count + 1, "Error Diff rules not working"
                        total_diff_count = count
                else:
                    assert count == i, "Error! Getting invalid count from cachedDiffs"
                    test_logger.debug("cachedDiffs Count Test passed successfully")
                    total_diff_count = count


if __name__ == "__main__":
    #test_commit_payloads()
    #test_diff_maps()
    test_diff_rules()
