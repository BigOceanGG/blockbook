#!/bin/sh

dir="/home/admin/coins/bcash/log"
if [ ! -d "$dir" ];then
  mkdir $dir
fi

FILE=~/coins/bcash/log/bcash_$(date "+%Y%m%d%H%M%S").log
./blockbook -workers=1 -sync -resyncindexperiod=200000 -blockchaincfg=build/bcash.json -datadir=/home/admin/coins/bcash/dbs -internal=:19002 -public=:19102 -logtostderr > ${FILE} 2>&1 &
rm ~/coins/bcash/bcash.log
ln -s ${FILE} ~/coins/bcash/bcash.log
