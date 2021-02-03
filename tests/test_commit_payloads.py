import requests
from random import choice
from string import ascii_letters
import time
import os

data = {
	"payload": {
		"constant": "DATA CONTSTANT",
		"message": "TEST",
		},
	"projectId": "dummy_init",
}

EXTERNAL_URL = os.getenv("EXTERNAL_IP", "localhost")

for i in range(5):
	message = ''.join([choice(ascii_letters) for i in range(10)])
	data['payload']['message'] = message
	print('*'*20, f' Try: {i} ', '*'*20)
	try:
		out = requests.post(url=f'http://{EXTERNAL_URL}:9000/commit_payload', json=data)
	except Exception as e:
		print(f"Failed {i}: {e}")
	else:
		print(f"Success {i}: {out.text}")
	print("*"*40)
	time.sleep(5)


