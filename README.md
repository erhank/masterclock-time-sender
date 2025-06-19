# UDP Time Sender

Sends current time as UDP multicast datagrams to 239.252.0.0:6168 every second.

## Usage

```bash
# Build everything
./control.sh build

# Send time packets
./control.sh send

# Listen for packets  
./control.sh listen

# Test packet format
./control.sh test
```

## Files

- `main.go` - Time sender application
- `listener.go` - Packet listener and analyzer
- `test.go` - Packet format test
- `control.sh` - Management script

## Packet Format

48-byte UDP payload with headers, device ID, and H:M:S time values.
See source code for detailed field layout.
