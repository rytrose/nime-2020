#!/bin/bash

docker-compose build server
docker tag nime-2020_server gcr.io/nime-2020/nime2020_server
docker push gcr.io/nime-2020/nime2020_server