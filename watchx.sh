#!/bin/sh

PRO_NAME=bitcoin.json
export LD_LIBRARY_PATH=./

while true; do
        NUM=`ps aux | grep ${PRO_NAME} | grep -v grep |wc -l`
        if [ "${NUM}" -lt "1" ]; then
            echo "$(date "+%Y-%m-%d %H:%M:%S") ${PRO_NAME} was killed"
            bash bcash.sh
            sleep 30
        fi

        MEM=`free | awk '/Mem/ {print $7}'`
        echo "$(date "+%Y-%m-%d %H:%M:%S") watch ${PRO_NAME} ${NUM} ${MEM}"
        sleep 30

done