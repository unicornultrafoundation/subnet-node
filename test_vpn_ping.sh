#!/bin/bash

# Configuration
TARGET_IP="10.0.0.3"
TOTAL_PINGS=1000
CONCURRENT_PINGS=50
TEST_DURATION=30
PING_INTERVAL=0.1

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Initialize counters
successful_pings=0
failed_pings=0
total_time=0
min_time=999999
max_time=0

# Create temporary file for results
temp_file=$(mktemp)
trap "rm -f $temp_file" EXIT

echo -e "${YELLOW}Starting VPN Ping Stress Test${NC}"
echo "Target IP: $TARGET_IP"
echo "Total Pings: $TOTAL_PINGS"
echo "Concurrent Pings: $CONCURRENT_PINGS"
echo "----------------------------------------"

# Function to perform a single ping and record results
do_ping() {
    local start_time=$(date +%s.%N)
    if ping -c 1 -W 1 $TARGET_IP >/dev/null 2>&1; then
        local end_time=$(date +%s.%N)
        local time_taken=$(echo "$end_time - $start_time" | bc)
        echo "SUCCESS $time_taken" >> "$temp_file"
    else
        echo "FAIL 0" >> "$temp_file"
    fi
}

# Start time of the test
start_total=$(date +%s.%N)

# Launch pings in parallel with rate limiting
for ((i=1; i<=TOTAL_PINGS; i++)); do
    # Check if we've exceeded the concurrent limit
    while [ $(jobs -r | wc -l) -ge $CONCURRENT_PINGS ]; do
        sleep 0.1
    done
    
    # Show progress
    if [ $((i % 50)) -eq 0 ]; then
        echo -ne "\rProgress: $i/$TOTAL_PINGS pings sent"
    fi
    
    # Launch ping in background
    do_ping &
    
    # Rate limiting
    sleep $PING_INTERVAL
done

# Wait for all pings to complete
wait
echo -e "\n\nAll pings completed. Processing results..."

# Process results
while IFS=' ' read -r result time; do
    if [ "$result" = "SUCCESS" ]; then
        successful_pings=$((successful_pings + 1))
        total_time=$(echo "$total_time + $time" | bc)
        
        # Update min/max times
        if (( $(echo "$time < $min_time" | bc -l) )); then
            min_time=$time
        fi
        if (( $(echo "$time > $max_time" | bc -l) )); then
            max_time=$time
        fi
    else
        failed_pings=$((failed_pings + 1))
    fi
done < "$temp_file"

# Calculate final statistics
end_total=$(date +%s.%N)
total_duration=$(echo "$end_total - $start_total" | bc)
avg_time=$(echo "scale=3; $total_time / $successful_pings" | bc 2>/dev/null)
success_rate=$(echo "scale=2; $successful_pings * 100 / $TOTAL_PINGS" | bc)
pings_per_second=$(echo "scale=2; $TOTAL_PINGS / $total_duration" | bc)

# Print results
echo -e "\n${YELLOW}Test Results:${NC}"
echo -e "----------------------------------------"
echo -e "Total Duration: ${GREEN}$(printf "%.2f" $total_duration) seconds${NC}"
echo -e "Successful Pings: ${GREEN}$successful_pings${NC}"
echo -e "Failed Pings: ${RED}$failed_pings${NC}"
echo -e "Success Rate: ${GREEN}${success_rate}%${NC}"
echo -e "Average Response Time: ${GREEN}$(printf "%.3f" $avg_time) seconds${NC}"
echo -e "Min Response Time: ${GREEN}$(printf "%.3f" $min_time) seconds${NC}"
echo -e "Max Response Time: ${GREEN}$(printf "%.3f" $max_time) seconds${NC}"
echo -e "Pings Per Second: ${GREEN}${pings_per_second}${NC}"

# Check memory usage of the VPN process
echo -e "\n${YELLOW}VPN Process Statistics:${NC}"
echo -e "----------------------------------------"
ps aux | grep "[s]ubnet" | awk '{printf "PID: %s\nCPU: %s%%\nMemory: %s%%\nCommand: %s\n", $2, $3, $4, $11}'

# Check system resources
echo -e "\n${YELLOW}System Resource Usage:${NC}"
echo -e "----------------------------------------"
free -m | awk 'NR==2{printf "Memory Usage: %s/%sMB (%.2f%%)\n", $3,$2,$3*100/$2 }'
df -h / | awk 'NR==2{printf "Disk Usage: %s/%s (%s)\n", $3,$2,$5}'
top -bn1 | grep "Cpu(s)" | sed "s/.,[0-9]*//" | awk '{print "CPU Usage:" $2"%"}'

# Final assessment
echo -e "\n${YELLOW}Test Assessment:${NC}"
echo -e "----------------------------------------"
if [ "$success_rate" -lt 90 ]; then
    echo -e "${RED}Warning: Success rate below 90%${NC}"
fi
if [ $(echo "$avg_time > 0.5" | bc -l) -eq 1 ]; then
    echo -e "${RED}Warning: High average response time${NC}"
fi
if [ "$failed_pings" -gt 0 ]; then
    echo -e "${RED}Warning: Some pings failed${NC}"
fi
if [ $(echo "$max_time > 1.0" | bc -l) -eq 1 ]; then
    echo -e "${RED}Warning: Some responses took more than 1 second${NC}"
fi