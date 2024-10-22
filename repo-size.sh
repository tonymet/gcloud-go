#!/bin/bash

HASHES_SIZES=$(git verify-pack -v .git/objects/pack/pack-*.idx \
  | sort -k 3 -rn \
  | head -20 \
  | awk '{print $1,$3}' \
  | sort)
HASHES=$(echo "$HASHES_SIZES" \
  | awk '{printf $1"|"}' \
  | sed 's/\|$//')
HASHES_FILES=$(git rev-list --objects --all \
  | \grep -E "($HASHES)" \
  | sort)
paste <(echo "$HASHES_SIZES") <(echo "$HASHES_FILES") \
  | sort -k 2 -rn \
  | awk '{
      size=$2; $1="";
      $2="";
      $3="";
      split( "KB MB GB" , v );
      s=0;
      while( size>1024 ){
        size/=1024; s++
      } print int(size) v[s]"\t"$0
    }'
  | column -ts $'\t'