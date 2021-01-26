#!/bin/bash

podman run \
    --rm \
    --name postgresql-nonpersistent \
    -p 5432:5432 \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres \
    -v postgresql-persistent:/var/lib/postgresql/data \
    $@ \
    postgres:13