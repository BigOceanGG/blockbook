./blockbook -workers=8 -sync -blockchaincfg=build/bitcoin.json -datadir=/home/admin/coins/bitcoin/dbs -internal=:19000 -public=:19100 -logtostderr > ~/coins/bitcoin/bitcoin.log 2>&1 &
