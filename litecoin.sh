#!/bin/sh

FILE=~/coins/litecoin/litecoin_$(date "+%Y%m%d%H%M%S").log
./blockbook -workers=1 -sync -blockchaincfg=build/litecoin.json -datadir=/home/admin/coins/litecoin/dbs -internal=:19004 -public=:19104 -logtostderr > ${FILE} 2>&1 &
rm ~/coins/litecoin/litecoin.log
ln -s ${FILE} ~/coins/litecoin/litecoin.log