./blockbook -workers=8 -sync -blockchaincfg=build/bitcoin.json -datadir=~/coins/dbs/bitcoin -internal=:19000 -public=:19100 -logtostderr > ~/coins/logs/bitcoin.log 2>&1 &
