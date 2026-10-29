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
- **Persistence** — append-only **AOF** with startup replay, live appends after writes,
  and periodic **RDB snapshots** with interval-based rewrites and size-based compaction.
  Snapshots record an AOF checkpoint so startup only replays the portion added after
  the snapshot.

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
./sider --checkpoint-file ./sider.checkpoint
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
| `INCR` | `INCR key` | Increments the integer stored at `key` by one; creates it as `1` if missing | `(integer) 2` |
| `TTL` | `TTL key` | Seconds until expiry; `-1` = forever, `-2` = missing | `(integer) 58` |
| `DEL` | `DEL key [key ...]` | Removes one or more keys; returns count removed | `(integer) 1` |
| `EXPIRE` | `EXPIRE key seconds` | Sets (or updates) the TTL of a key | `(integer) 1` |

Reply types follow Redis: `GET` always returns a **bulk string**, while `INCR`
returns an **integer**. `INCR` only accepts base-10 integer values — applying it
to a float or non-numeric string returns an error and leaves the value unchanged.

## Architecture

```
cmd/sider/
└── main.go          # entry point: flags → server bootstrap

config/
└── config.go        # default host (0.0.0.0), port (7030), AOF path and RDB snapshot path

internal/
├── server/
│   ├── server.go        # shared read/respond helpers + sync TCP server
│   ├── epoll_linux.go   # async epoll event loop (Linux only)
│   ├── epoll_other.go   # non-Linux stub for the async server
│   └── conn.go          # FDComm — socket Read/Write over a raw fd (Linux)
├── command/
│   ├── command.go       # RedisCmd / RedisCmds — parsed command representation
│   ├── dispatcher.go    # command dispatch (EvalAndRespond)
│   ├── string.go        # PING, SET, GET, INCR handlers
│   ├── keyspace.go      # TTL, DEL, EXPIRE handlers
│   └── admin.go         # COMMAND, BGREWRITEAOF handlers (AOF rewrite)
├── store/
│   ├── store.go         # in-memory key/value store: Value + ExpiresAt per key
│   ├── expiration.go    # active sampled expiry cleanup
│   └── eviction.go      # (stub) future LRU / LFU / random eviction
├── persistence/
│   ├── aof.go           # AOF append / rewrite / replay logic
│   └── rdb.go           # RDB snapshot: take / load / schedule
└── resp/
    ├── decoder.go       # RESP2 decoder: strings, errors, ints, bulk, arrays, null
    ├── encoder.go       # RESP2 encoder for the same types
    └── constants.go     # shared RESP reply constants (OK, PONG, NIL, ...)
```

### How the pieces fit together

1. `main.go` parses the server and persistence flags and starts `server.RunTCPAsyncServer()`.
2. The server creates a non-blocking listening socket, wraps it in an **epoll**
   instance, and waits with `syscall.EpollWait`.
3. Read-ready sockets are demultiplexed: the listener gets `Accept`ed, and each
   client fd is added to the epoll interest set as an `FDComm`.
4. Incoming bytes are decoded with the **RESP2 decoder** into one or more
   `RedisCmd`s (this is what enables pipelining).
5. `command.EvalAndRespond` dispatches each command through the matching `Eval*`
   handler, encodes the response, and writes it straight back to the socket.
6. When epoll finds **nothing to do** (idle), the active expiry `store.Cleanup()` runs.

### Expiry: passive + active

- **Passive** — `GET` (and friends) check `ExpiresAt` on access; an expired key is
  deleted on the spot and reported as missing.
- **Active** — every idle event-loop turn, `Cleanup()` samples up to 20 expiring
  keys; if more than 25% of the sample has expired, it keeps sweeping until the
  expired ratio drops below that threshold.

## Roadmap

- [x] Persistence (RDB snapshots)
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