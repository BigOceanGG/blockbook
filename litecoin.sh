./blockbook -workers=8 -sync -blockchaincfg=build/litecoin.json -datadir=/home/admin/coins/litecoin/dbs -internal=:19004 -public=:19104 -logtostderr > ~/coins/litecoin/litecoin_$(date "+%Y%m%d%H%M%S").log 2>&1 &
