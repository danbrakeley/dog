#!/bin/bash
set -e

##
## show differences between logger/level/fields between dog and frog
##

trap 'rm -rf "$TMPDIR"' EXIT
TMPDIR=$(mktemp -d) || exit 1

for file in fields.go level.go level_test.go logger.go null.go null_test.go
do
  echo "Comparing $file..."
  curl https://raw.githubusercontent.com/danbrakeley/frog/main/$file --no-progress-meter -o $TMPDIR/$file
  sed -e 's/^package frog$/package dog/' < "$TMPDIR/$file" > "$TMPDIR/frog.$file"
  diff -u $TMPDIR/frog.$file ./$file || true
done

echo "Done"
