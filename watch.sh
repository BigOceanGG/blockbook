#!/bin/sh

export LD_LIBRARY_PATH=./

while true; do

        NUM=`free | awk '/Mem/ {print $7}'`
        if [ ${NUM} -lt 1000000 ]; then
            pkill -2 blockbook
            echo "$(date "+%Y-%m-%d %H:%M:%S") pkill -2 ... ${NUM}"
            sleep 6
            pkill -9 blockbook
            echo "$(date "+%Y-%m-%d %H:%M:%S") pkill -9 ... ${NUM}"
            sleep 6
            bash bitcoin.sh
            echo "$(date "+%Y-%m-%d %H:%M:%S") start ..."
       else
            echo "$(date "+%Y-%m-%d %H:%M:%S") watch ... ${NUM}"
        fi
        sleep 30

done