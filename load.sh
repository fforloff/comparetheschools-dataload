#!/bin/bash

for year in 2016
#for year in {2006..2016}
do
echo $year
./dataload -database "gotestdb" -in ~/code/comparetheschools/raw_data/*${year}*.csv  -year $year
done
