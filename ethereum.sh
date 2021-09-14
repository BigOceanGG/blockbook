./blockbook -workers=8 -sync -blockchaincfg=build/ethereum.json -datadir=~/coins/ethereum/dbs -internal=:19003 -public=:19103 -logtostderr > ~/coins/logs/ethereum.log 2>&1 &
