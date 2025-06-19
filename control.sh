#!/bin/bash

echo "UDP Time Sender - Control Script"
echo "================================"
echo "1. Build all applications"
echo "2. Run time sender"
echo "3. Run packet listener (in background)"
echo "4. Run both (sender + listener)"
echo "5. Stop all"
echo "6. Show packet analysis"
echo ""

case $1 in
  "build"|"1")
    echo "Building applications..."
    go build -o time-sender main.go
    go build -o packet-listener listener.go
    go build -o packet-test test.go packet.go
    echo "Built: time-sender, packet-listener, packet-test"
    ;;
  "send"|"2")
    echo "Starting time sender..."
    ./time-sender
    ;;
  "listen"|"3")
    echo "Starting packet listener..."
    ./packet-listener
    ;;
  "both"|"4")
    echo "Starting both sender and listener..."
    echo "Listener output will be in listener.log"
    ./packet-listener > listener.log 2>&1 &
    LISTENER_PID=$!
    echo "Listener PID: $LISTENER_PID"
    sleep 2
    echo "Starting sender..."
    ./time-sender
    ;;
  "stop"|"5")
    echo "Stopping all processes..."
    pkill -f "time-sender"
    pkill -f "packet-listener"
    echo "All processes stopped"
    ;;
  "test"|"6")
    echo "Running packet format test..."
    ./packet-test
    ;;
  *)
    echo "Usage: $0 {build|send|listen|both|stop|test}"
    echo "  or: $0 {1|2|3|4|5|6}"
    ;;
esac
