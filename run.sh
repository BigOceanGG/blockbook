#!/bin/sh

FILE=watch_$(date "+%Y%m%d%H%M%S")

nohup bash watch.sh >${FILE}.log &
rm watch.log
ln -s ${FILE}.log watch.log