#!/bin/bash

ADDR="http://127.0.0.1:10101"

curl $ADDR/index/users -X POST

# 性别
curl $ADDR/index/users/field/gender -X POST -d '{"options": {"type": "mutex"}}'

pilosa import --host=$ADDR -i users --field gender csv/gender.csv

# 阅读
curl $ADDR/index/users/field/read_day -X POST -d '{"options": {"type": "time", "timeQuantum": "YMD"}}'
curl $ADDR/index/users/field/read_book_count -X POST -d '{"options": {"type": "mutex"}}'

pilosa import --host=$ADDR -i users --field read_day csv/read_day.csv
pilosa import --host=$ADDR -i users --field read_book_count csv/read_book_count.csv

# 下载
curl $ADDR/index/users/field/down_day -X POST -d '{"options": {"type": "time", "timeQuantum": "YMD"}}'
curl $ADDR/index/users/field/down_book_count -X POST -d '{"options": {"type": "mutex"}}'

pilosa import --host=$ADDR -i users --field down_day csv/down_day.csv
pilosa import --host=$ADDR -i users --field down_book_count csv/down_book_count.csv

# 加入书架
curl $ADDR/index/users/field/shelf_day -X POST -d '{"options": {"type": "time", "timeQuantum": "YMD"}}'
curl $ADDR/index/users/field/shelf_book_count -X POST -d '{"options": {"type": "mutex"}}'

pilosa import --host=$ADDR -i users --field shelf_day csv/shelf_day.csv
pilosa import --host=$ADDR -i users --field shelf_book_count csv/shelf_book_count.csv


