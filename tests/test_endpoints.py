import requests
from string import ascii_letters
from random import choice
import time


def test_commit_payloads():
    project_id = ''.join([choice(ascii_letters) for i in range(10)])
    for i in range(10):
        for j in range(5)
            data = {'data':'data_'+str(i+1)}
            payload = {'payload':data, 'projectId':'some_'}
            out = requests.post('http://localhost:9000/commit_payload', json=payload)
            assert out.status_code == 200, f"Recieved {out.status_code} from /commit_payload endpoint"
            data = out.json()
            keys = list(data.keys())
            for k in ['cid','tentativeHeight', 'payloadChanged']:
                assert k in keys, "Invalid response recieved from the /commit_payload endpoint" 
            
            assert data['tentativeHeight']  == (i+1)*(j+1), "Invalid tentativeHeight from /commit_payload endpoint response"
            
            if (j == 0) and (i != 0):
                assert data['payloadChanged'] == "true", "Invalid payload_changed flag recieved from /commit_payload endpoint"
            else:
                assert data['payloadChanged'] == "false", "Invalid payload_changed flag recieved from /commit_payload endpoint"


def test_getBlocks():
    """ First test for gettings single blocks that are on the """
