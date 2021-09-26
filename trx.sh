./blockbook -workers=128 -sync -blockchaincfg=build/trx.json -datadir=/home/admin/coins/trx/dbs -internal=:19001 -public=:19101 -logtostderr > ~/coins/trx/trx_$(date "+%Y%m%d%H%M%S").log 2>&1 &
