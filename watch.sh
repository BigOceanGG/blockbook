#!/bin/sh

export LD_LIBRARY_PATH=./

while true; do

        NUM=`free | awk '/Mem/ {print $7}'`
        if [ ${NUM} -lt 1000000 ]; then
            PID=`ps aux | grep ethereum.json| grep -v grep|awk '{print $2}'`
            if  [ $PID ]; then
              echo "$(date "+%Y-%m-%d %H:%M:%S") ethereum was killed"
              kill -2 ${PID}
              echo "$(date "+%Y-%m-%d %H:%M:%S") kill -2 ... ${NUM}"
              sleep 120
              kill -9 ${PID}
              echo "$(date "+%Y-%m-%d %H:%M:%S") kill -9 ... ${NUM}"
              sleep 10
              bash ethereum.sh
              echo "$(date "+%Y-%m-%d %H:%M:%S") start ..."
            fi
       else
            echo "$(date "+%Y-%m-%d %H:%M:%S") watch ... ${NUM}"
        fi
        sleep 30

done