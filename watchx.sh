#!/bin/sh

PRO_NAME=blockbook
export LD_LIBRARY_PATH=./

while true; do
        NUM=`ps aux | grep ${PRO_NAME} | grep -v grep |wc -l`
        if [ "${NUM}" -lt "1" ]; then
            echo "${PRO_NAME} was killed"
            bash bitcoin.sh
            sleep 30
        fi

        MEM=`free | awk '/Mem/ {print $7}'`
        echo "watch ${PRO_NAME} ${NUM}"
        sleep 60

done