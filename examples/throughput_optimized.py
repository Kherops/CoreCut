#!/usr/bin/env python3
"""
Example optimized for throughput mode.
Outputs: THROUGHPUT: <value>
"""
import time
import json
from concurrent.futures import ThreadPoolExecutor

def process_item(item):
    """Same processing logic"""
    result = 0
    for i in range(1000):
        result += i * item
    return result

def main():
    start = time.time()
    
    items = list(range(1000))
    
    # Optimized: parallel processing with thread pool
    with ThreadPoolExecutor(max_workers=4) as executor:
        results = list(executor.map(process_item, items))
    
    # Optimized: batch serialization
    _ = json.dumps([{"item": i, "result": r} for i, r in zip(items, results)])
    
    elapsed = time.time() - start
    throughput = len(items) / elapsed
    
    print(f"THROUGHPUT: {throughput:.2f}")

if __name__ == "__main__":
    main()
