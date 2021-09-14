./blockbook -workers=8 -sync -blockchaincfg=build/trx.json -datadir=~/coins/dbs/trx -internal=:19001 -public=:19101 -logtostderr > ~/coins/logs/trx.log 2>&1 &
