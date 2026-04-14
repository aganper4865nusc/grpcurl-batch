# grpcurl-batch

A CLI tool for batch-executing gRPC calls from a YAML manifest with retry and concurrency controls.

---

## Installation

```bash
go install github.com/yourusername/grpcurl-batch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/grpcurl-batch.git
cd grpcurl-batch
go build -o grpcurl-batch .
```

---

## Usage

Define your gRPC calls in a YAML manifest:

```yaml
# manifest.yaml
host: localhost:50051
concurrency: 4
calls:
  - method: mypackage.MyService/GetUser
    data: '{"id": "123"}'
    retries: 3
  - method: mypackage.MyService/ListItems
    data: '{"page": 1}'
    retries: 1
```

Then run:

```bash
grpcurl-batch --manifest manifest.yaml
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--manifest` | `manifest.yaml` | Path to the YAML manifest file |
| `--concurrency` | `1` | Number of concurrent gRPC calls |
| `--timeout` | `30s` | Timeout per individual call |
| `--plaintext` | `false` | Use plaintext (no TLS) |

---

## Requirements

- Go 1.21+
- A running gRPC server with reflection enabled (or a `.proto` file)

---

## License

This project is licensed under the [MIT License](LICENSE).