#!/usr/bin/env bash

# Build the container
# Bend ID / route
docker build -f Dockerfile -t g6w0qh/foo:latest .


# Start the container
docker run -v $(pwd)/workspace:/bridge --env-file=env --rm --pull=never  g6w0qh/foo:latest

# Debug the container
# docker run -v $(pwd)/workspace:/bridge --rm --pull=never -it --entrypoint /bin/sh g6w0qh/foo:latest