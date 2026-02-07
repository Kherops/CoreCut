#!/bin/bash
# Example optimized script - same work but with optimizations
# Uses batching, reduced syscalls, and better I/O patterns

WORKDIR=$(mktemp -d)
trap "rm -rf $WORKDIR" EXIT

# Optimized: batch file creation with single write
{
    for i in {1..100}; do
        echo "Processing item $i"
    done
} > "$WORKDIR/batch_output.txt"

# Optimized: use arithmetic expansion instead of external commands
for i in {1..100}; do
    # Inline computation (no subprocess)
    : $((i * i))
done

# Optimized: single read of batch file
cat "$WORKDIR/batch_output.txt" > /dev/null

# Optimized: use bash arithmetic instead of seq + subshell
sum=0
for ((n=1; n<=10000; n++)); do
    ((sum += n * n))
done

echo "Optimized complete"
