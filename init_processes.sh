#!/bin/bash

echo 'starting pm2...';

pm2 start pm2.config.js

pm2 logs --lines 1000
