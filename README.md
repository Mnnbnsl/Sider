# SIDER

**A minimal, Redis-compatible in-memory data store built from scratch in Go.**

SIDER is my personal, from-the-ground-up understanding of how Redis works — the wire
protocol, the event loop, the expiry machinery, and the command engine — implemented
by hand with almost no dependencies outside the standard library.

> **Note:** This is a learning project. It is not a drop-in replacement for Redis and
> is not intended for production use.

---

## Highlights

- **RESP2 wire protocol** — a hand-written decoder and encoder for Simple Strings,
  Errors, Integers, Bulk Strings, Arrays, and null values.
- **Epoll-based async server** — a Linux epoll event loop that handles thousands of
  concurrent client connections from a single thread.
- **Command pipelining** — multiple commands sent in a single write are parsed and
  answered in one round-trip.
- **TTL & expiry** — per-key expiry with **passive** cleanup (on access) and **active**
  sampled cleanup (during idle event-loop cycles).

## Getting Started

### Prerequisites

- Go **1.25** or later
- Linux (the async server uses kernel epoll via `syscall`)

### Build & Run

```bash
git clone <https://github.com/Mnnbnsl/sider> && cd sider
go build -o sider ./cmd/sider

./sider                          # listens on 0.0.0.0:7030
./sider --port 6379             # custom port
./sider --host 127.0.0.1        # bind only to loopback
```

You should see:

```
2026/09/16 12:00:00 Gearing up SIDER...
2026/09/16 12:00:00 starting an asynchronous TCP server on 0.0.0.0 7030
```

### Try it with redis-cli

Any Redis client works out of the box — `redis-cli` is the easiest.

```bash
$ redis-cli -p 7030 ping
PONG

$ redis-cli -p 7030 set greeting "hello sider"
OK

$ redis-cli -p 7030 get greeting
"hello sider"

$ redis-cli -p 7030 expire greeting 60
(integer) 1

$ redis-cli -p 7030 ttl greeting
(integer) 58

$ redis-cli -p 7030 del greeting
(integer) 1
```

Looking inside the box: what is actually sent and received on the wire is the RESP2
protocol, e.g. `SET greeting "hello sider"` becomes:

```
*3\r\n$3\r\nSET\r\n$8\r\ngreeting\r\n$12\r\nhello sider\r\n
```

and SIDER replies with `+OK\r\n`.

## Supported Commands

| Command | Signature | Behavior | Example reply |
| --- | --- | --- | --- |
| `PING` | `PING [message]` | Returns `PONG`, or echoes `message` if given | `PONG` / `hello` |
| `SET` | `SET key value [EX seconds]` | Stores a value, optionally expiring after `EX` seconds | `OK` |
| `GET` | `GET key` | Retrieves a value, or null if missing/expired | `"value"` / `(nil)` |
| `TTL` | `TTL key` | Seconds until expiry; `-1` = forever, `-2` = missing | `(integer) 58` |
| `DEL` | `DEL key [key ...]` | Removes one or more keys; returns count removed | `(integer) 1` |
| `EXPIRE` | `EXPIRE key seconds` | Sets (or updates) the TTL of a key | `(integer) 1` |

## Architecture

```
cmd/sider/
└── main.go          # entry point: flags → server bootstrap

config/
└── main.go          # default host (0.0.0.0) and port (7030)

internal/
├── server/
│   └── server.go    # async epoll event loop + sync TCP server,
│                    #   multi-command (pipelined) parsing
├── core/
│   ├── store.go     # in-memory key/value store: Value + ExpiresAt per key
│   ├── cmd.go       # RedisCmd / RedisCmds — parsed command representation
│   ├── eval.go      # command dispatch & all command handlers (Eval*)
│   ├── expiration.go# active sampled expiry cleanup
│   ├── eviction.go  # (stub) future LRU / LFU / random eviction
│   └── comm.go      # FDComm — socket Read/Write over a raw file descriptor
└── resp/
    ├── decoder.go   # RESP2 decoder: strings, errors, ints, bulk, arrays, null
    └── encoder.go   # RESP2 encoder for the same types
```

### How the pieces fit together

1. `main.go` parses `--host` / `--port` and starts `server.RunTCPAsyncServer()`.
2. The server creates a non-blocking listening socket, wraps it in an **epoll**
   instance, and waits with `syscall.EpollWait`.
3. Read-ready sockets are demultiplexed: the listener gets `Accept`ed, and each
   client fd is added to the epoll interest set as an `FDComm`.
4. Incoming bytes are decoded with the **RESP2 decoder** into one or more
   `RedisCmd`s (this is what enables pipelining).
5. `core.EvalAndRespond` dispatches each command through the matching `Eval*`
   handler, encodes the response, and writes it straight back to the socket.
6. When epoll finds **nothing to do** (idle), the active expiry `Cleanup()` runs.

### Expiry: passive + active

- **Passive** — `GET` (and friends) check `ExpiresAt` on access; an expired key is
  deleted on the spot and reported as missing.
- **Active** — every idle event-loop turn, `Cleanup()` samples up to 20 expiring
  keys; if more than 25% of the sample has expired, it keeps sweeping until the
  expired ratio drops below that threshold.

## Roadmap

- [ ] Persistence (RDB snapshots / AOF append-only file)
- [ ] Eviction policies — LRU, LFU, and random, for `allkeys-*` and `volatile-*`
- [ ] Streaming reads (clients currently read in fixed 512-byte chunks)
- [ ] RESP3 protocol features
- [ ] More data types (lists, hashes, sets, sorted sets)
- [ ] Automated test suite
- [ ] Bloom Filters , Hyperloglog etc.

## Environment & Dependencies

| | |
| --- | --- |
| Language | Go 1.25 |
| Dependencies | `golang.org/x/sys` (raw fd I/O & epoll) — everything else is stdlib |
| Platforms | Linux (epoll); macOS/Windows would need a kqueue/IOCP backend |

## Acknowledgements

Shout-out to the real [Redis](https://redis.io) — the source of inspiration and the
reference for both the RESP2 protocol and the general server architecture. Reading
Redis's codebase and the RESP spec is highly recommended.