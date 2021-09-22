#!/bin/sh

PRO_NAME = bitcoin.json
export LD_LIBRARY_PATH = ./

while true; do
        NUM = `ps aux | grep ${PRO_NAME} | grep -v grep |wc -l`
        if [ "${NUM}" -lt "1" ]; then
            echo "${PRO_NAME} was killed"
            ./blockbook -workers=1 -sync -blockchaincfg=build/bitcoin.json -datadir=/home/admin/coins/bitcoin/dbs -internal=:19000 -public=:19100 -logtostderr > ~/coins/bitcoin/bitcoin.log 2>&1 &
            sleep 10
        fi

        echo "watch ${PRO_NAME}"
        sleep 60

done