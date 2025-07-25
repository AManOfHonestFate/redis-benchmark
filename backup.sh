#!/bin/bash
set -e

BACKUP_DIR=./backup
mkdir -p $BACKUP_DIR

date=$(date +%Y%m%d_%H%M%S)
for d in ./data/node-*; do
  node=$(basename $d)
  tar czf $BACKUP_DIR/${node}_$date.tar.gz -C $d .
done

echo "Backups saved to $BACKUP_DIR"

