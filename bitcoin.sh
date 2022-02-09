#!/bin/sh

dir="/home/ubuntu/btc/log"
if [ ! -d "$dir" ];then
  mkdir $dir
fi

FILE=~/btc/log/bitcoin_$(date "+%Y%m%d%H%M%S").log
./blockbook -workers=1 -sync -resyncindexperiod=300000 -blockchaincfg=build/bitcoin.json -datadir=/home/ubuntu/btc/dbs -internal=:19000 -public=:19100 -logtostderr > ${FILE} 2>&1 &
rm ~/btc/bitcoin.log
ln -s ${FILE} ~/btc/bitcoin.log