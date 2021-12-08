#!/bin/sh

dir="/home/ubuntu/bitcoin-0.21.1/log"
if [ ! -d "$dir" ];then
  mkdir $dir
fi

FILE=~/bitcoin-0.21.1/log/bitcoin_$(date "+%Y%m%d%H%M%S").log
./blockbook -workers=1 -sync -resyncindexperiod=300000 -blockchaincfg=build/bitcoin.json -datadir=/home/ubuntu/bitcoin-0.21.1/dbs -internal=:19000 -public=:19100 -logtostderr > ${FILE} 2>&1 &
rm ~/bitcoin-0.21.1/bitcoin.log
ln -s ${FILE} ~/bitcoin-0.21.1/bitcoin.log