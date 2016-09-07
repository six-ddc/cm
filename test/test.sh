#!/bin/bash

REDIS_HOST="127.0.0.1"
REDIS_PORT="6379"

REDIS_PREFIX="cm_test"

_test() {
    echo "-----------------"
    cat example.ini
    echo "-----------------"

    cm --redis "${REDIS_HOST}:${REDIS_PORT}" --redis-prefix "${REDIS_PREFIX}" set db.passwd=12345 db.username=ddc

    key="${REDIS_PREFIX}:db.username"
    echo "key:${key}"
    redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} hgetall $key

    key="${REDIS_PREFIX}:db.passwd"
    echo "key:${key}"
    redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} hgetall $key

    cm --redis "${REDIS_HOST}:${REDIS_PORT}" --redis-prefix "${REDIS_PREFIX}" get example.ini

    echo "-----------------"
    cat example.ini.out
    echo "-----------------"
}

_test
