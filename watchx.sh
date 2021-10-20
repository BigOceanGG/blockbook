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

        echo "watch ${PRO_NAME}"
        sleep 60

done