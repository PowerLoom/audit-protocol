#!/bin/bash

cd docker/

if [ "$1" == "pg" ]
then
  make powerloom-pg-build
else
  make powerloom-build
fi
