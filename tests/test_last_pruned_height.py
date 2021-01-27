from utils import helper_functions
import asyncio


async def test_func():
	out = await helper_functions.get_last_pruned_height('dummy_init')
	print(out)

if __name__ == "__main__":
	try:
		asyncio.get_event_loop().run_until_complete(asyncio.gather(test_func()))
	except:
		pass
	finally:
		asyncio.get_event_loop().stop()


