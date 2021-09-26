./blockbook -workers=8 -sync -blockchaincfg=build/bcash.json -datadir=/home/admin/coins/bcash/dbs -internal=:19002 -public=:19102 -logtostderr > ~/coins/bcash/bcash_$(date "+%Y%m%d%H%M%S").log 2>&1 &
