#!/bin/bash
# Example baseline script - simulates a slow process
# This does sequential file operations without optimization

WORKDIR=$(mktemp -d)
trap "rm -rf $WORKDIR" EXIT

# Simulate slow sequential I/O
for i in {1..100}; do
    echo "Processing item $i" > "$WORKDIR/file_$i.txt"
    # Simulate some CPU work
    for j in {1..1000}; do
        echo "$j" > /dev/null
    done
    # Read back (no caching benefit)
    cat "$WORKDIR/file_$i.txt" > /dev/null
done

# Simulate some computation
seq 1 10000 | while read n; do
    echo $((n * n)) > /dev/null
done

echo "Baseline complete"
