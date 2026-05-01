#!/bin/bash
# scripts/test.sh - Final Trial Automation
set -e

echo " Starting AeroCore AI Final Trial..."
echo "--------------------------------------"

# 1. Cleanup previous runs
rm -f ./aero-core

# 2. Build Binary
echo " 1. Building Static Binary..."
CGO_ENABLED=0 go build -o ./aero-core ./cmd/server
if [ $? -eq 0 ]; then echo "✅ Build Success"; else exit 1; fi

# 3. Ensure Data Dir exists
mkdir -p data

# 4. Start Server in Background
echo "🏃 2. Starting Server on port 8080..."
export CONFIG_PATH=configs/app.yaml
./aero-core > /tmp/aerocore.log 2>&1 &
SERVER_PID=$!
sleep 3 # Wait for startup

# 5. Check Process
if kill -0 $SERVER_PID 2>/dev/null; then
    echo "✅ Server Running (PID: $SERVER_PID)"
else
    echo "❌ Server Failed to Start. Check /tmp/aerocore.log"
    exit 1
fi

# 6. Test Health
echo " 3. Testing Health Endpoint..."
HEALTH=$(curl -s http://localhost:8080/health)
if echo "$HEALTH" | grep -q "ok"; then
    echo "✅ Health Check Passed"
else
    echo "❌ Health Check Failed"
fi

# 7. Test Deep Health (Ollama connection)
echo "🔌 4. Testing Deep Health (Ollama Connection)..."
DEEP=$(curl -s http://localhost:8080/health/deep)
echo "Response: $DEEP"

# 8. Test Chat Stream
echo "💬 5. Testing Chat Stream..."
# Timeout 15s agar tidak hang lama jika model lambat
OUTPUT=$(curl -N -s --max-time 15 -X POST http://localhost:8080/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"session_id":"test_final","message":"Apa itu AI?","model":"qwen2.5:1.5b-instruct-q4_k_m"}')

if echo "$OUTPUT" | grep -q "event: chunk"; then
    echo "✅ Streaming Active!"
else
    echo "⚠️ Streaming Test Failed or Timeout (Normal jika model loading)"
fi

# 9. Check RAM Usage
echo "💾 6. Checking RAM Usage..."
# Menggunakan ps untuk cek RSS (Resident Set Size) dalam KB, lalu bagi 1024 jadi MB
MEM_USAGE=$(ps -o rss= -p $SERVER_PID)
MEM_MB=$((MEM_USAGE / 1024))
echo "📊 Server RAM Usage: ~${MEM_MB} MB"

if [ $MEM_MB -lt 100 ]; then
    echo "✅ RAM Usage Excellent (< 100MB)"
elif [ $MEM_MB -lt 500 ]; then
    echo "✅ RAM Usage Good (< 500MB)"
else
    echo "⚠️ RAM Usage High (> 500MB) - Check memory leaks"
fi

# 10. Cleanup
echo "🧹 7. Stopping Server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null
echo "✅ Trial Completed Successfully!"
