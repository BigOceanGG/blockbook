#!/bin/sh

FILE=~/coins/bcash/bcash_$(date "+%Y%m%d%H%M%S").log
./blockbook -workers=1 -sync -blockchaincfg=build/bcash.json -datadir=/home/admin/coins/bcash/dbs -internal=:19002 -public=:19102 -logtostderr > ${FILE} 2>&1 &
rm ~/coins/bcash/bcash.log
ln -s ${FILE} ~/coins/bcash/bcash.log
