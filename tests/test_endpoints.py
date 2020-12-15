import requests
import time
from test_logger import test_logger


def test_commit_payloads():
    project_id = "PROJECT_COMMIT_PAYLOAD"
    test_logger.debug("Using the project Id: ")
    test_logger.debug(project_id)
    for i in range(10):
        for j in range(5):
            test_logger.debug("Using i and j: {}, {}, height: {}".format(i, j, (i * 5) + (j + 1)))
            data = {'data': 'data_' + str(i + 1), 'constant_string': "CONSTANT_STRING_HOLDER"}
            payload = {'payload': data, 'projectId': project_id}
            test_logger.debug("Committing payload: ")
            test_logger.debug(payload)
            out = requests.post('http://localhost:9000/commit_payload', json=payload)
            assert out.status_code == 200, f"Received {out.status_code} from /commit_payload endpoint"
            data = out.json()
            test_logger.debug("Got Response: ")
            test_logger.debug(data)
            keys = list(data.keys())
            for k in ['cid', 'tentativeHeight', 'payloadChanged']:
                assert k in keys, "Invalid response received from the /commit_payload endpoint"

            assert data['tentativeHeight'] == (i * 5) + (j + 1), \
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
            old_height = requests.get(f'http://localhost:9000/{project_id}/payloads/height')
            old_height = old_height.json()['height']
            test_logger.debug("Using i and j: {}, {}, old_height: {}".format(i, j, old_height))
            data = {'data': 'data_' + str(i + 1), 'constant_string': "CONSTANT_STRING_HOLDER"}
            payload = {'payload': data, 'projectId': project_id}
            test_logger.debug("Committing payload: ")
            test_logger.debug(payload)
            out = requests.post('http://localhost:9000/commit_payload', json=payload)
            assert out.status_code == 200, f"Received {out.status_code} from /commit_payload endpoint"
            data = out.json()
            test_logger.debug("Got Response: ")
            test_logger.debug(data)
            while True:
                new_height = requests.get(f'http://localhost:9000/{project_id}/payloads/height')
                new_height = new_height.json()['height']
                if new_height == old_height:
                    test_logger.debug("Chain has not been updated: {}, {}".format(old_height, new_height))
                    time.sleep(3)
                else:
                    test_logger.debug("Chain has been updated: {}, {}".format(old_height, new_height))
                    break
            if (j == 0) and (i != 0):
                test_logger.debug("Getting the cachedDiff count from the audit-protocol")
                out = requests.get(f"http://localhost:9000/{project_id}/payloads/cachedDiffs/count")
                out = out.json()
                test_logger.debug("Got response:")
                test_logger.debug(out)

                count = out['count']
                assert count == i, "Error! Getting invalid count from cachedDiffs"





if __name__ == "__main__":
    test_commit_payloads()
    test_diff_maps()
