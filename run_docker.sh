#!/bin/bash

maticvigil_settings_json=$HOME/.maticvigil/settings.json
maticvigil_account_info_json=$HOME/.maticvigil/account_info.json


if [ -f "$maticvigil_settings_json" ] && [ -f "$maticvigil_account_info_json" ]
then
	echo "Found maticvigil files"
else
	echo "Error: No mactivigil files found. Make sure you have settings.json and account_info.json in your $HOME/.maticvigil directory"
	exit 1
fi


cp "$maticvigil_settings_json" docker/
cp "$maticvigil_account_info_json" docker/

cd docker/
make powerloom-build