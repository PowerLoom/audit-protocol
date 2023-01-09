#!/bin/bash

MATICVIGIL_SETTINGS_JSON="${HOME}/.maticvigil/settings.json"
MATICVIGIL_ACCOUNT_INFO_JSON="${HOME}/.maticvigil/account_info.json"


if [ -f $MATICVIGIL_SETTINGS_JSON ] && [ -f $MATICVIGIL_ACCOUNT_INFO_JSON ]
then
	echo "Found maticvigil files"
else
	echo "Error: No mactivigil files found. Make sure you have settings.json and account_info.json in your $HOME/.maticvigil directory"
	#exit 1
fi


cp "$MATICVIGIL_SETTINGS_JSON" docker/
cp "$MATICVIGIL_ACCOUNT_INFO_JSON" docker/

cd docker/

if [ "$1" == "pg" ]
then
  make powerloom-pg-build
else
  make powerloom-build
fi
