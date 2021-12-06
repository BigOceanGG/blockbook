#!/bin/sh

FILE=~/coins/bitcoin/bitcoin_$(date "+%Y%m%d%H%M%S").log
./blockbook -workers=1 -sync -resyncindexperiod=300000 -blockchaincfg=build/bitcoin.json -datadir=/home/ubuntu/bitcoin-22.0/dbs -internal=:19000 -public=:19100 -logtostderr > ${FILE} 2>&1 &
rm ~/bitcoin-22.0/bitcoin.log
ln -s ${FILE} ~/bitcoin-22.0/bitcoin.log