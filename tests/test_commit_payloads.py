import requests
from random import choice
from string import ascii_letters
import time

data = {
	"payload": {
		"constant": "DATA CONTSTANT",
		"message": "TEST",
		},
	"projectId": "dummy_init",
}

for i in range(30):
	message = ''.join([choice(ascii_letters) for i in range(10)])
	data['payload']['message'] = message
	print('*'*20,f' Try: {i} ', '*'*20)
	try:
		out = requests.post(url='http://localhost:9000/commit_payload', json=data)
	except Exception as e:
		print(f"Failed {i}: {e}")
	else:
		print(f"Success {i}: {out.text}")
	print("*"*40)
	time.sleep(0.5)


