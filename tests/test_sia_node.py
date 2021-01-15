from utils import *
from random import choice
from string import ascii_letters
import hashlib
import asyncio
import requests

async def test_sia_node():
	random_text = ''.join([choice(ascii_letters) for i in range(10000)])
	print(random_text)
	t_h = hashlib.md5(random_text.encode()).hexdigest()
	
	await sia_upload(t_h, random_text)

	print("Getting the file from sia")
	out = requests.get(url=f"http://localhost:9980/renter/stream/{t_h}",headers={'user-agent': 'Sia-Agent'}, stream=True)
	print(out)
	for data in out.iter_content(chunk_size=1024*50):
		print(data)


if __name__ == "__main__":
	asyncio.get_event_loop().run_until_complete(test_sia_node())
