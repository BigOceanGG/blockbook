#!/bin/sh

FILE=~/coins/dogecoin/dogecoin_$(date "+%Y%m%d%H%M%S").log
./blockbook -workers=1 -sync -resyncIndexPeriodMs=30000 -blockchaincfg=build/dogecoin.json -datadir=/home/admin/coins/dogecoin/dbs -internal=:19005 -public=:19105 -logtostderr > ${FILE} 2>&1 &
rm ~/coins/dogecoin/dogecoin.log
ln -s ${FILE} ~/coins/dogecoin/dogecoin.log