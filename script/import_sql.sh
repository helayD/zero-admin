#!/bin/bash
BASE="/mnt/c/Users/qianmai88/Documents/GitHub/zero-admin/script/sql"
MYSQL_CMD="docker exec -i zero-mysql mysql -uroot -p12341qweqfsd2356 --default-character-set=utf8mb4 gozero"

for dir in sys ums cms oms pms sms pub; do
  for f in "$BASE/$dir"/*.sql; do
    echo "Importing: $f"
    $MYSQL_CMD < "$f" 2>&1 || true
  done
done
echo "=== DONE ==="
