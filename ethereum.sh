#!/bin/sh

FILE=~/coins/ethereum/ethereum_$(date "+%Y%m%d%H%M%S").log
./blockbook -workers=128 -sync -blockchaincfg=build/ethereum.json -datadir=/home/admin/coins/ethereum/dbs -internal=:19003 -public=:19103 -logtostderr > ${FILE} 2>&1 &
rm ~/coins/ethereum/ethereum.log
ln -s ${FILE} ~/coins/ethereum/ethereum.log