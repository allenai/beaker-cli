#!/bin/sh

set -e

# This script:
# 1. lists all files under `/input` with `find`
# 2. extracts each file's contents with `cat`
# 3. counts the words/lines in the resulting text with `wc`

if [ -z "$COUNT_LINES" ]; then

  echo "Counting words..."
  COUNT=$(find /input -type f -print0 | xargs -0 cat | wc -w)
  echo "{\"word_count\": $COUNT}" > /output/metrics.json

else

  echo "Counting lines..."
  COUNT=$(find /input -type f -print0 | xargs -0 cat | wc -l)
  echo "{\"line_count\": $COUNT}" > /output/metrics.json

fi

echo "Done!"
