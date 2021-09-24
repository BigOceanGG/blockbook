#!/bin/sh

export LD_LIBRARY_PATH=./

while true; do

        NUM=`free | awk '/Mem/ {print $7}'`
        if [ ${NUM} -lt 600000 ]; then
            pkill -2 blockbook
            echo "$(date "+%Y-%m-%d %H:%M:%S") pkill -2 ... ${NUM}"
            sleep 60
            pkill -9 blockbook
            echo "$(date "+%Y-%m-%d %H:%M:%S") pkill -9 ... ${NUM}"
            sleep 5
            bash trx.sh
            bash ethereum.sh
            echo "$(date "+%Y-%m-%d %H:%M:%S") start ..."
        else
            echo "$(date "+%Y-%m-%d %H:%M:%S") watch ... ${NUM}"
        fi
        sleep 120

done