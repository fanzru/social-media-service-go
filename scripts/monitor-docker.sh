#!/bin/bash

# Simple Docker stats monitor - Send to InfluxDB
# Run this in background: ./scripts/monitor-docker.sh &

while true; do
    docker stats --no-stream --format "{{.Name}}\t{{.CPUPerc}}\t{{.MemPerc}}" 2>/dev/null | while IFS=$'\t' read name cpu mem; do
        # Remove % symbol
        cpu_value=$(echo $cpu | sed 's/%//')
        mem_value=$(echo $mem | sed 's/%//')
        
        # Send to InfluxDB
        curl -s -X POST "http://localhost:8086/api/v2/write?org=social-media&bucket=metrics" \
          -H "Authorization: Token my-super-secret-auth-token" \
          -H "Content-Type: text/plain; charset=utf-8" \
          -d "docker_stats,container=$name cpu_percent=$cpu_value,memory_percent=$mem_value" > /dev/null 2>&1
    done
    sleep 5
done


