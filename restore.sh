#!/bin/bash
set -e
BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
  echo "Usage: $0 <backup_file>"
  exit 1
fi

NODE_DIR=$(echo $BACKUP_FILE | sed -E 's/.*(node-[0-9]+)_.*/\1/')

if [ ! -d ./data/$NODE_DIR ]; then
  mkdir -p ./data/$NODE_DIR
fi

tar xzf $BACKUP_FILE -C ./data/$NODE_DIR
echo "Restored $BACKUP_FILE to ./data/$NODE_DIR"

