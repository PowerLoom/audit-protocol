from random import choice
from string import ascii_letters
import hashlib
import asyncio
import requests
import sys

def load_file(filename):
	try:
		f = open(filename,'r')
		data = f.read()
	except Exception as e:
		print("There was an error while loading the file: ")
		print(e)

		return -1
	return data


async def test_sia_node():
	filename = sys.argv[1]
	data = load_file(filename)
	if data == -1:
		print("Invalid file provided")
		sys.exit(0)

	n_bytes = sys.getsizeof(data)
	print("Size of data in MB: ")
	print(n_bytes/(10**6))

	t_h = hashlib.md5(data.encode()).hexdigest()

	#data = {'data': random_text}	
	#await sia_upload(t_h, random_text)
	headers = {'user-agent': 'Sia-Agent', 'content-type':'application/octet-stream'}
	try:
		out = requests.post(
				url=f"http://localhost:9980/renter/uploadstream/{t_h}?datapieces=10&paritypieces=20",
				headers=headers,
				data=data,
			)
		print(out)
		print(out.text)
	except Exception as e:
		print("There was an error while pushing the file to Sia: ")
		print(e)
		return -1

	print("Getting the file from sia")
	try:
		out = requests.get(url=f"http://localhost:9980/renter/stream/{t_h}", headers={'user-agent': 'Sia-Agent'}, stream=True)
	except Exception as e:
		print("There was an error while getting the file: ")
		print(e)
		return -1
	print(out)
	print(out.text)
	for data in out.iter_content(chunk_size=1024*50):
		print(data)


if __name__ == "__main__":
	if len(sys.argv) < 2:
		print("Enter the filename of the file that needs to be pushed to sia")
		sys.exit(0)
	loop = asyncio.get_event_loop()
	try:
		loop.run_until_complete(test_sia_node())
	except Exception as e:
		print("An error occure while running the test")
		print(e)
	finally:
		loop.stop()
