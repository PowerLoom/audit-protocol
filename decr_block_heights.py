import redis

# Get all projectId's
redis_client = redis.Redis()

all_projects = redis_client.keys('projectID*blockHeight')

backup_heights = {}
for key in all_projects:
	backup_heights[key.decode('utf-8')] = redis_client.get(key).decode('utf-8')
	redis_client.decr(key)


import json
backup_heights_files = open("backup_height.json","w")
json.dump(backup_heights, backup_heights_files)
