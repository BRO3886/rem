#!/bin/bash
# Benchmark script for rem CLI commands
# Runs each command 3 times and reports min/avg/max times

set -e

BINARY="bin/rem"

if [ ! -f "$BINARY" ]; then
    echo "Binary not found. Run 'make build' first."
    exit 1
fi

benchmark() {
    local desc="$1"
    shift
    local cmd="$@"
    local times=()

    for i in 1 2 3; do
        # Use /usr/bin/time for precision
        local t=$( { time $cmd > /dev/null 2>&1; } 2>&1 )
        local real=$(echo "$t" | grep real | awk '{print $2}')
        # Convert to seconds
        local mins=$(echo "$real" | sed 's/m.*//')
        local secs=$(echo "$real" | sed 's/.*m//' | sed 's/s//')
        local total=$(echo "$mins * 60 + $secs" | bc)
        times+=("$total")
    done

    # Calculate min/avg/max
    local min=$(printf '%s\n' "${times[@]}" | sort -n | head -1)
    local max=$(printf '%s\n' "${times[@]}" | sort -n | tail -1)
    local sum=0
    for t in "${times[@]}"; do
        sum=$(echo "$sum + $t" | bc)
    done
    local avg=$(echo "scale=3; $sum / 3" | bc)

    printf "%-40s min=%-8s avg=%-8s max=%-8s\n" "$desc" "${min}s" "${avg}s" "${max}s"
}

echo "=== rem CLI Benchmark (3 runs each) ==="
echo "Date: $(date)"
echo "Binary: $BINARY"
echo ""

benchmark "version"                    $BINARY version
benchmark "lists"                      $BINARY lists
benchmark "lists --count"              $BINARY lists --count
benchmark "list (all reminders)"       $BINARY list
benchmark "list --list Inbox"          $BINARY list --list Inbox
benchmark "list --incomplete"          $BINARY list --list Inbox --incomplete
benchmark "show (by prefix)"           $BINARY show 9679EF91
benchmark "search omscs"               $BINARY search omscs
benchmark "stats"                      $BINARY stats
benchmark "overdue"                    $BINARY overdue
benchmark "upcoming"                   $BINARY upcoming
benchmark "export --format json"       $BINARY export --format json
benchmark "export --format csv"        $BINARY export --format csv

echo ""
echo "=== Done ==="
