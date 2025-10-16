#!/usr/bin/env python3
"""
Performance testing script for Nexoan Query API
Runs the curl command 20 times and calculates average response times
python3 perf/performance_test.py --count 10
"""

import subprocess
import re
import statistics
import time
import json
import os
import argparse

def run_curl_test(url, debug=False):
    """Run a single curl test and parse the timing results"""
    
    curl_command = [
        'curl', '-w', '\nnamelookup:%{time_namelookup} connect:%{time_connect} appconnect:%{time_appconnect} starttransfer:%{time_starttransfer} total:%{time_total}\n',
        '-o', '/dev/null', '-s',
        '-X', 'POST',
        url,
        '-H', 'Content-Type: application/json',
        '-d', json.dumps({
            "id": "",
            "kind": {
                "major": "Organisation",
                "minor": "minister"
            },
            "name": "",
            "created": "",
            "terminated": ""
        })
    ]
    
    try:
        result = subprocess.run(curl_command, capture_output=True, text=True, timeout=30)
        
        if result.returncode != 0:
            print(f"Error: curl command failed with return code {result.returncode}")
            print(f"Error output: {result.stderr}")
            return None
        
        # Debug mode: show full response
        if debug:
            print(f"\nDebug - Full response: {result.stdout}")
            print(f"Debug - Return code: {result.returncode}")
        
        # Check if we have timing data
        has_timing_data = any(line.startswith(('namelookup:', 'connect:', 'appconnect:', 'starttransfer:', 'total:')) for line in result.stdout.split('\n'))
        
        if not has_timing_data:
            # This might be an HTTP error response, let's check
            if '404' in result.stdout or 'Not Found' in result.stdout:
                if debug:
                    print(f"HTTP 404 Error: {result.stdout.strip()}")
                return None
            elif '500' in result.stdout or 'Internal Server Error' in result.stdout:
                if debug:
                    print(f"Server Error: {result.stdout.strip()}")
                return None
            else:
                if debug:
                    print(f"Unexpected response: {result.stdout.strip()}")
                return None
            
        # Parse the timing output
        timing_data = {}
        lines = result.stdout.strip().split('\n')
        
        for line in lines:
            if line.strip():
                # Split by spaces to handle multiple timing values on one line
                parts = line.split()
                for part in parts:
                    if ':' in part:
                        key, value = part.split(':', 1)
                        key = key.strip()
                        value = value.strip()
                        try:
                            timing_data[key] = float(value)
                        except ValueError:
                            if debug:
                                print(f"Debug - Could not parse timing value: '{key}': '{value}'")
                            continue
        
        if debug:
            print(f"Debug - Parsed timing data: {timing_data}")
            
        # Check if we got at least the total time
        if 'total' not in timing_data:
            if debug:
                print("Debug - No 'total' timing found")
            return None
                    
        return timing_data
        
    except subprocess.TimeoutExpired:
        print("Error: Request timed out")
        return None
    except Exception as e:
        print(f"Error running curl: {e}")
        return None

def main():
    """Run 20 tests and calculate averages"""
    # Parse command line arguments
    parser = argparse.ArgumentParser(description='Performance test for Nexoan Query API')
    parser.add_argument('--url', '-u', 
                       default=os.getenv('QUERY_API_URL', 'https://aaf8ece1-3077-4a52-ab05-183a424f6d93-prod.e1-us-east-azure.choreoapis.dev/data-platform/query-api/v1.1'),
                       help='Query API URL (default: from QUERY_API_URL env var or hardcoded default)')
    parser.add_argument('--count', '-c', type=int, default=20,
                       help='Number of requests to run (default: 20)')
    parser.add_argument('--delay', '-d', type=float, default=0.5,
                       help='Delay between requests in seconds (default: 0.5)')
    parser.add_argument('--debug', action='store_true',
                       help='Enable debug mode to show full responses')
    
    args = parser.parse_args()
    
    print("🚀 Starting Nexoan Query API Performance Test")
    print("=" * 60)
    print(f"Target URL: {args.url}")
    print(f"Running {args.count} requests to measure performance...")
    print()
    
    # Store all timing results
    all_results = []
    successful_tests = 0
    
    for i in range(args.count):
        print(f"Test {i+1:2d}/{args.count}: ", end="", flush=True)
        
        result = run_curl_test(args.url, debug=args.debug)
        
        if result:
            all_results.append(result)
            successful_tests += 1
            print(f"✅ Total: {result.get('total', 0):.3f}s")
        else:
            print("❌ Failed")
        
        # Small delay between requests
        time.sleep(args.delay)
    
    print()
    print("=" * 60)
    print(f"📊 Performance Test Results ({successful_tests}/{args.count} successful)")
    print("=" * 60)
    
    if successful_tests == 0:
        print("❌ No successful tests completed")
        return
    
    # Calculate statistics for each timing metric
    metrics = ['namelookup', 'connect', 'appconnect', 'starttransfer', 'total']
    
    print(f"{'Metric':<15} {'Min':<8} {'Max':<8} {'Avg':<8} {'Median':<8} {'Std Dev':<8}")
    print("-" * 65)
    
    for metric in metrics:
        values = [result.get(metric, 0) for result in all_results if metric in result]
        
        if values:
            min_val = min(values)
            max_val = max(values)
            avg_val = statistics.mean(values)
            median_val = statistics.median(values)
            std_dev = statistics.stdev(values) if len(values) > 1 else 0
            
            print(f"{metric:<15} {min_val:<8.3f} {max_val:<8.3f} {avg_val:<8.3f} {median_val:<8.3f} {std_dev:<8.3f}")
    
    print()
    print("📈 Summary:")
    print(f"  • Successful requests: {successful_tests}/{args.count} ({successful_tests/args.count*100:.1f}%)")
    
    if 'total' in all_results[0]:
        total_times = [result['total'] for result in all_results]
        avg_total = statistics.mean(total_times)
        min_total = min(total_times)
        max_total = max(total_times)
        
        print(f"  • Average total time: {avg_total:.3f}s")
        print(f"  • Fastest request: {min_total:.3f}s")
        print(f"  • Slowest request: {max_total:.3f}s")
        
        # Performance assessment
        if avg_total < 1.0:
            print("  • Performance: 🟢 Excellent (< 1s)")
        elif avg_total < 2.0:
            print("  • Performance: 🟡 Good (1-2s)")
        elif avg_total < 5.0:
            print("  • Performance: 🟠 Acceptable (2-5s)")
        else:
            print("  • Performance: 🔴 Slow (> 5s)")

if __name__ == "__main__":
    main()
