#!/bin/sh

FILE=~/coins/trx/trx_$(date "+%Y%m%d%H%M%S").log
./blockbook -workers=512 -sync -blockchaincfg=build/trx.json -datadir=/home/admin/coins/trx/dbs -internal=:19001 -public=:19101 -logtostderr > ${FILE} 2>&1 &
rm ~/coins/trx/trx.log
ln -s ${FILE} ~/coins/trx/trx.log