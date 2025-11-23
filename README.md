# uinputd-go

A Linux daemon for keyboard input automation with multi-layout support.

Inspired by [ydotool](https://github.com/ReimuNotMoe/ydotool).

## Overview

`uinputd` creates a virtual keyboard device and listens for input automation commands via a Unix socket. It enables secure, script-friendly keyboard input emulation with support for multiple keyboard layouts.

## Features

- **Virtual Keyboard Device**: Uses Linux `/dev/uinput` for native input emulation
- **Multi-Layout Support**: US, FR, DE, ES, UK, IT keyboard layouts
- **Unix Socket IPC**: JSON-based protocol for client-daemon communication
- **Real-time Streaming**: Character-by-character typing with configurable delays
- **Batch Typing**: Fast text input automation
- **Security Hardening**: Systemd sandboxing and group-based access control
- **Low Resource Usage**: Efficient concurrent request handling
- **Programmatic API**: Go client library for integration

## Architecture

```
uinputd-go/
├── cmd/
│   ├── uinputd/          # Daemon binary
│   └── uinput-client/    # CLI client
├── internal/
│   ├── config/           # Configuration management
│   ├── logger/           # Structured logging
│   ├── uinput/           # Virtual keyboard device
│   ├── layouts/          # Keyboard layout implementations
│   ├── protocol/         # Command/response messages
│   └── server/           # Unix socket server
├── pkg/
│   └── client/           # Public Go client library
├── configs/              # Configuration templates
└── systemd/              # Systemd service unit
```

## Quick Start

### Build and Install

```bash
# Build the client (includes embedded daemon)
make build

# Install the client
sudo make install
```

### Install Daemon and Service

The client includes embedded installation commands:

```bash
# Install the daemon
sudo uinput-client install daemon

# Install systemd service
sudo uinput-client install systemd-service

# Enable and start the service
sudo systemctl enable uinputd
sudo systemctl start uinputd

# Add your user to the input group (required for client access)
sudo usermod -aG input $USER
# (logout and login for group changes to take effect)
```

### Usage

**Type text:**
```bash
uinput-client type "Hello, World!"
```

**Type with specific layout:**
```bash
uinput-client type "Bonjour le monde" --layout fr
```

**Stream text (real-time typing):**
```bash
uinput-client stream "Typing in real-time..."
```

**Press a key:**
```bash
uinput-client key KEY_ENTER
uinput-client key KEY_A --modifier shift
```

**Health check:**
```bash
uinput-client ping
```

## Configuration

Default config locations (in order of priority):
1. `--config` flag
2. `/etc/uinputd/uinputd.yaml`
3. `~/.config/uinputd/uinputd.yaml`
4. `./uinputd.yaml`

Environment variables: `UINPUTD_*` (e.g., `UINPUTD_SOCKET_PATH`)

### Example Configuration

```yaml
socket:
  path: /tmp/.uinputd.sock
  permissions: 0600

layout: us

performance:
  buffer_size: 4096
  max_message_size: 1048576
  stream_delay_ms: 50
  char_delay_ms: 10
  max_concurrent_cmds: 100

logging:
  level: info
  format: auto
```

## Supported Layouts

- `us` - US QWERTY
- `fr` - French AZERTY
- `de` - German QWERTZ
- `es` - Spanish
- `uk` - UK QWERTY
- `it` - Italian

## Programmatic Usage

```go
import "github.com/yourusername/uinputd-go/pkg/client"

c, err := client.New("/tmp/.uinputd.sock")
if err != nil {
    log.Fatal(err)
}
defer c.Close()

// Type text
err = c.TypeText(ctx, "Hello!", "us")

// Stream text
err = c.StreamText(ctx, "Real-time typing", "us", 50, 10)

// Press a key
err = c.SendKey(ctx, "KEY_ENTER", "")
```

## Requirements

- Linux kernel with uinput support
- Root privileges (for `/dev/uinput` access)
- Go 1.25.3+ (for building)

## Security

- Daemon runs as root (required for `/dev/uinput`)
- Socket permissions: `0660` with `root:input` group
- Systemd sandboxing: `NoNewPrivileges`, `ProtectSystem`, `ProtectHome`
- Local-only communication via Unix socket

## Build Targets

```bash
make build              # Build daemon + client
make install            # Install client to /usr/local/bin
make test               # Run tests
make test-coverage      # Generate coverage report
make clean              # Clean build artifacts
make help               # Display all targets
```

## Development

**Run daemon manually:**
```bash
sudo ./bin/uinputd --config configs/uinputd.yaml
```

**Test client:**
```bash
./bin/uinput-client type "Test message"
```

## License

See LICENSE file for details.
