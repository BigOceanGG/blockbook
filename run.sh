#!/bin/sh

if [ $# -eq 0 ]; then
  echo "para error"
  exit 1
fi


dir="log"
if [ ! -d "$dir" ];then
  mkdir $dir
fi

FILE=watch_$(date "+%Y%m%d%H%M%S")

nohup bash watchx.sh $1 > log/${FILE}.log &
rm watch.log
ln -s log/${FILE}.log watch.log