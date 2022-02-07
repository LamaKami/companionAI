#!/bin/bash
docker build -t "$1" .
docker run -v $(pwd):/app -u ${UID} -d -p 5000:5000 "$1"