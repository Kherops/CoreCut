#!/usr/bin/env python3
"""
Example baseline for throughput mode.
Outputs: THROUGHPUT: <value>
"""
import time
import json

def process_item(item):
    """Simulate slow processing - no batching"""
    result = 0
    for i in range(1000):
        result += i * item
    return result

def main():
    start = time.time()
    
    items = list(range(1000))
    results = []
    
    # Sequential processing - no optimization
    for item in items:
        result = process_item(item)
        results.append(result)
        # Simulate I/O per item
        _ = json.dumps({"item": item, "result": result})
    
    elapsed = time.time() - start
    throughput = len(items) / elapsed
    
    print(f"THROUGHPUT: {throughput:.2f}")

if __name__ == "__main__":
    main()
