#!/bin/sh

FILE=~/coins/bitcoin/bitcoin_$(date "+%Y%m%d%H%M%S").log
./blockbook -workers=128 -sync -blockchaincfg=build/bitcoin.json -datadir=/home/admin/coins/bitcoin/dbs -internal=:19000 -public=:19100 -logtostderr > ${FILE} 2>&1 &
rm ~/coins/bitcoin/bitcoin.log
ln -s ${FILE} ~/coins/bitcoin/bitcoin.log