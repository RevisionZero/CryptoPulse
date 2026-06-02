# NexusCorr — Development Notes

---

## Completed ✅

- **Welford's Algorithm** — running mean/variance with catastrophic cancellation protection
- **Circular/Ring Buffer** — fixed-size sliding window for price history
- **Circuit Breaker** — connection failure handling in the connector
- **Streams Creator** — takes a list of symbols and constructs the Binance endpoint
- **Map Pre-allocation in `CalculatePCCMatrix`** — matrix created once in Synchronizer and updated in-place rather than reallocated every tick
- **Channel-based Sampler/PriceUpdater** — replaced shared mutex with idiomatic Go channels for ownership-based concurrency
- **Separate Read Pump in Connector** — dedicated goroutine reads from the WebSocket and writes to an internal buffered channel; circuit breaker and processing logic run in a separate select loop, so `ReadMessage()` is never blocked by downstream work (`connector.go`)
- **Non-blocking send in `SendToAll`** — channel sends use `select`/`default` to skip full or closed client buffers without blocking the Hub or panicking (`hub.go:202–208`)
- **Connector goroutine lifecycle** — each symbol's connector goroutine receives a `closeChan` and exits when the last subscribed client disconnects; `RemoveSymbols` decrements the reference count and signals close when it reaches zero (`hub.go:155–169`)
- **Unused symbol cleanup on disconnect** — `RemoveClient` calls `RemoveSymbols`, which decrements per-symbol client counters and deletes the entry when it hits zero (`hub.go:145–169`)
- **`sync.Pool` for JSON buffer reuse** — a `bufferPool` recycles `bytes.Buffer` instances used during JSON encoding in `SendToAll`, reducing per-tick allocations (`hub.go:44–49`)
- **Structured logging** — `log/slog` with a JSON handler is configured in `main.go` and used consistently across all packages in place of `fmt.Println`
- **Allowed origins / `CheckOrigin`** — the WebSocket upgrader in `wsHandler.go` explicitly allows only `https://cryptopulseapp.dev` and `http://localhost:5173`, blocking all other origins
- **Error handling** — all major parsing and I/O paths have explicit error checks with log output: JSON unmarshal and float parsing in `priceUpdater.go`, dial errors in `connector.go`, JSON encode errors in `hub.go`, WebSocket upgrade errors in `wsHandler.go`
- **Empty-check guards** — `CalculatePCCMatrix` skips any symbol pair where either data slice is missing or empty (`pcc.go:47–49`)
- **Case normalization** — `endpointConstructor.go` lowercases all symbols when constructing the Binance endpoint URL; symbols are stored in uppercase matching Binance's response field `"s"`, so keys are consistent throughout
- **Frontend ticker validation** — the UI validates each ticker against the Binance REST API before sending to the WebSocket, and auto-appends `USDT` if a bare symbol (e.g. `BTC`) is entered (`App.tsx:97–118`)
- **Connection status indicator** — the UI displays a three-state animated badge: "LIVE DATA" (green), "WAITING FOR TICKERS" (orange), and "CONNECTION CLOSED" (red), with a pulsing dot for each state (`App.tsx:166–183`, `App.css:21–81`)
- **Server disconnect message** — the `close` event on the WebSocket sets the connection state to `'disconnected'`, surfacing the "CONNECTION CLOSED" badge to the user immediately
- **Correlation heatmap** — table cells use an inline `backgroundColor` computed from the correlation value: green for positive, red for negative, with opacity scaled by `Math.abs(value)` (`App.tsx:259–263`)
- **Informative tooltips** — a `?` icon next to the "Correlation Matrix" heading shows a plain-English explanation of PCC on hover (`App.tsx:228–236`)
- **User-selectable stream count (UI)** — the ticker input section exposes 5 fields, enforces a minimum of 2 non-empty tickers, and labels the first two as required (`App.tsx:194–222`)
- **Testing** — `pcc_test.go` contains three unit tests covering positive correlation, negative correlation, and no correlation; `engine_test.go` provides a signal-feeding integration harness

---

## Algorithm & Math

- **Inverse Welford** — implement incremental removal of data points from the running statistics, enabling a true sliding window without a full recalculation on every tick
- **Floating-point drift** — over days of operation, rounding errors in the Inverse logic accumulate; re-sync the full sum from scratch every 1,000 samples to keep it numerically accurate
- **Staleness detection** — add a check to detect when a symbol's data has gone dead (e.g., BTC streaming zeros for an extended period) and handle it explicitly rather than silently corrupting the correlation output
- **Signals and systems from sampling** — revisit sampling theory considerations from the 4X03 course in the context of the 100ms tick rate and Welford's numerical stability

---

## Backend — Connector & WebSocket

- **Non-blocking send to `dataChan` in Connector** — the send `dataChan <- msg.message` (`connector.go:62`) is currently a blocking operation; if the Hub is slow, the read pump stalls and Binance's pings go unanswered, causing a connection drop; wrap this in a `select`/`default` to shed load and log a warning instead of blocking
- **Hard exit for dial retries** — the retry loop (`for baseWait > 0`) never terminates because `baseWait` only grows; add a maximum retry count so the connector gives up and returns after a sustained outage rather than spinning forever
- **Error propagation from `connector.go`** — connection errors are currently logged and absorbed internally; decide explicitly whether they should bubble up to the caller (e.g., via a returned error channel) so the Hub can react to a permanently failed symbol
- **Backfill missing data after disconnect** — when a connector reconnects after a drop, the sliding window contains a gap; decide on an explicit strategy: skip the gap, flag the window as invalid for one full period, or interpolate
- **Resource cleanup on reconnect** — `conn.dial()` inside the retry loop overwrites `conn.conn` without closing the previous connection; call `conn.conn.Close()` before each re-dial to avoid leaking file descriptors

---

## Backend — Hub & Concurrency

- **Race condition on `hub.symbols`** — the per-symbol lock acquisition in `HandleSymbolRequest` is correct, but the lock is a `sync.Mutex`; since `PriceUpdater` and `Sampler` only read the map structure (they write into individual `SymbolAttributes`), upgrading to `sync.RWMutex` would allow reads to proceed concurrently with each other
- **Race condition at startup** — the Synchronizer begins ticking immediately, before any symbols are registered; while `CalculatePCCMatrix` guards against empty slices, verify there is no window where `Sampler` iterates the symbols map while `HandleSymbolRequest` is mid-insertion
- **More efficient locking in Sampler** — the lock is currently held for the entire loop body, including `SlidingWindow.GetAll()` which allocates and copies; take a snapshot of the relevant fields under the lock and release it before doing any computation
- **Worker pool for PCC matrix calculation** — instead of the Hub loop blocking on matrix math per client group, dispatch `MatrixJob` structs to a fixed pool of worker goroutines (sized to CPU core count); workers return finished JSON to a results channel for the Hub to broadcast, keeping the 100ms pulse on schedule
- **JSON pre-marshalling** — move `json.Marshal` out of `SendToAll` and into the `Run()` broadcast case before calling `SendToAll`; the Hub then distributes pre-calculated bytes, and the marshalling cost is paid once per tick regardless of client count (the `sync.Pool` already reduces allocation cost, but the marshal still happens per unique symbol group inside the hot path)
- **Pre-join symbol keys** — `strings.Join(client.Symbols, ",")` allocates a new string on every `SendToAll` call for every client; store the pre-joined `SymbolKey` string on the `Client` struct when the symbol request is processed so the inner loop is a simple pointer read
- **Ping/Pong heartbeats in `writePump`** — mobile browsers and corporate proxies close idle WebSocket connections after 30–60 seconds even when data is flowing; add a 30-second Ping ticker to `writePump` to keep connections alive and detect half-open sockets quickly
- **Graceful shutdown** — `main.go` catches `SIGINT` and logs "Interrupt received", but returns without closing any WebSocket connections or stopping goroutines; add explicit cleanup so clients receive a proper close frame rather than a hard disconnect
- **Locking entire hub for symbol addition request** — When a client requests a symbol, in handleSymbolRequest, a lock on the entire hub is done. Check if that's necessary

---

## Backend — Code Quality

- **Server-side input validation** — the backend splits the raw WebSocket message on commas and acts on it with no further checks (`wsHandler.go:80`); validate symbol format, count, and allowed characters before passing to `HandleSymbolRequest`
- **Custom unmarshaller** — `BinanceTicker` carries `EventType`, `TransTime`, and `EventTime` fields that are parsed but never used; replace with a minimal struct or a custom unmarshaller that only extracts `Symbol`, `BestBid`, and `BestAsk`
- **Reduce Sampler parameters** — `Sampler(symbols, symbolLock, sampledDataChan)` takes three parameters that all derive from the same Hub state; consider wrapping them in a struct or having the Synchronizer own them directly
- **Unexplained constants** — `CircuitBreaker{Closed, 9, 20, 0, 0, true}` in `connector.go` has magic numbers for `failThreshold` and `successNeeded`; document why 9 failures and 20 successes were chosen (at 100ms per tick, 20 successes = 2 seconds of stable data before reopening)
- **Documentation** — add package-level and function-level doc comments where the behavior or invariants are non-obvious; the circuit breaker state machine and the Sampler lock strategy are the most important candidates

---

## Frontend / UI

- **Replace `alert()`** — `handleValidateAndSubmit` still calls `alert('Please enter at least 2 tickers')` (`App.tsx:126`); replace with an inline validation message consistent with the per-field error display already in place
- **Data update animations** — the table has a `background-color` CSS transition (`App.css:128`) but no per-arrival pulse effect; adding a brief keyframe animation triggered on data change would make live updates visually obvious without being distracting
- **Relative measurements** — several values are hard-coded in pixels (`padding: 12px 16px`, `min-width: 100px`, `width: 140px`, `width: 8px`, etc.); convert to `rem` or `%` where appropriate for consistent scaling across screen densities
- **Responsive/mobile layout** — no media queries exist; the correlation table overflows on small screens (the `overflow-x: auto` on `.table-container` helps but is not sufficient); add a stacked list view for viewports below a breakpoint

---

## Features

- **User-selectable correlation window** — allow the user to choose the lookback period (e.g., last 5 minutes, last 100 data points) rather than using the fixed 600-sample window hard-coded in `synchronizer.go`
- **Dynamic ticker management** — support removing individual symbols at runtime without a full reconnect; the backend already handles connector lifecycle correctly, but there is no UI or protocol for a client to desubscribe from a specific symbol
- **Enforce 2–5 ticker limit server-side** — the UI enforces the minimum of 2, but the backend applies no cap; a malformed or crafted request could subscribe to an arbitrary number of symbols

---

## Deployment

- **Docker** — no `Dockerfile` or `docker-compose.yml` exists in the repository yet; containerize the backend with a minimal base image
- **Caddy** — no `Caddyfile` exists yet; configure as the reverse proxy with automatic HTTPS, WebSocket proxying (`reverse_proxy localhost:8080`), and appropriate read/write timeouts
- **Oracle Cloud — VCN Security Rules** — open inbound TCP on ports 80, 443, and 8080 via the OCI Console under Networking → VCN → Security Lists
- **Oracle Cloud — OS-level firewall** — run the following to open the ports in `iptables` and persist across reboots:
  ```bash
  sudo iptables -I INPUT 6 -p tcp --dport 80 -j ACCEPT
  sudo iptables -I INPUT 6 -p tcp --dport 443 -j ACCEPT
  sudo iptables -I INPUT 6 -p tcp --dport 8080 -j ACCEPT
  sudo netfilter-persistent save
  ```
- **Domain setup** — point the domain to the Oracle Cloud instance IP and verify DNS propagation before enabling Caddy's HTTPS provisioning

---

## Other

- **Set up project management and kanban board** — Set up kanban board with different issue types, etc.

---

## Useful Commands

| Action | Command |
|---|---|
| SSH into server | `ssh -i ssh-key-2026-01-07.key ubuntu@40.233.69.209` |
| Disconnect | `exit` |
| Stream backend logs | `sudo docker compose logs -f app` |
