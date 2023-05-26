#!/bin/sh

set -e

SOURCE=""
IPFS_DOCKER_CONTAINER_NAME=""

# read flags from command line
while getopts ":s:n:" opt; do
  case $opt in
    s)
      SOURCE=$OPTARG
      echo "SOURCE: $SOURCE"
      ;;
    n)
      IPFS_DOCKER_CONTAINER_NAME=$OPTARG
      echo "local ipfs docker container name: $IPFS_DOCKER_CONTAINER_NAME"
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      ;;
  esac
done

echo "create state dump on source instance"
# login into pooler docker instance
# create state dump  = poetry run python -m pooler.protocol_state_loader_exporter export
# copy state dump to host
POOLER_CONTAINER_ID=$(ssh "$SOURCE" docker ps -aqf name=deploy-pooler-1)
ssh "$SOURCE" << EOF
cd pooler
#docker exec -i deploy-pooler-1 bash -c "poetry run python -m pooler.protocol_state_loader_exporter export"
docker cp $POOLER_CONTAINER_ID:/state.json.bz2 .
EOF

echo "create ipfs cache archive on source instance"
# get source instance's ipfs docker container id
# copy ipfs cache to host
# archive the cache folder
# remove the cache folder
IPFS_CONTAINER_ID=$(ssh "$SOURCE" docker ps -aqf name=deploy-ipfs-1)
ssh "$SOURCE" << EOF
docker cp $IPFS_CONTAINER_ID:/data/ipfs .
tar -cf ipfs.tar ipfs
rm -r ipfs
EOF

STATE_DUMP_PATH="/root/pooler/state.json.bz2"
IPFS_STATE_DUMP_PATH="/root/ipfs.tar"
LOCAL_CACHE_DUMP_PATH="/root/local_cache.tar"

echo "fetching local cache dump from $SOURCE at path $LOCAL_CACHE_DUMP_PATH"
scp "$SOURCE":"$LOCAL_CACHE_DUMP_PATH" .

echo "unzipping local cache dump archive"
tar -xvf local_cache.tar
cd local_cache && cp -r * /tmp && cd ../ && rm -rf local_cache

echo "removing local cache dump archive"
rm local_cache.tar

echo "fetching state dump from $SOURCE at path $STATE_DUMP_PATH"
scp "$SOURCE":"$STATE_DUMP_PATH" .

# get file name from state dump path
archivename=$(basename -- "$STATE_DUMP_PATH")
echo "unzipping state dump archive $archivename"
gunzip "$archivename"
cp state.json /tmp/state.json && rm state.json

echo "fetching ipfs state dump, might take a while"
scp "$SOURCE":"$IPFS_STATE_DUMP_PATH" .

echo "unzipping ipfs state dump archive"
# get file name from ipfs dump path
archivename=$(basename -- "$IPFS_STATE_DUMP_PATH")
tar -xvf "$archivename"

echo "removing ipfs state dump archive"
rm "$archivename"

echo "fetching ipfs docker container id"
IPFS_CONTAINER_ID=$(docker ps -aqf "name=$IPFS_DOCKER_CONTAINER_NAME")

echo "copying ipfs state dump to docker container"
docker cp ipfs "$IPFS_CONTAINER_ID":/data/

echo "removing ipfs state dump"
rm -rf ipfs

echo "restarting ipfs docker container"
docker restart "$IPFS_CONTAINER_ID"

echo "waiting for ipfs docker container to restart, sleeping for 10 seconds"
sleep 10

echo "finished"
