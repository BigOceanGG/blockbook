#!/bin/sh

FILE=~/coins/bitcoin/bitcoin_$(date "+%Y%m%d%H%M%S").log
./blockbook -workers=1 -sync -resyncindexperiod=300000 -blockchaincfg=build/bitcoin.json -datadir=/home/admin/coins/bitcoin/dbs -internal=:19000 -public=:19100 -dbcache=1073741824 -logtostderr > ${FILE} 2>&1 &
rm ~/coins/bitcoin/bitcoin.log
ln -s ${FILE} ~/coins/bitcoin/bitcoin.log