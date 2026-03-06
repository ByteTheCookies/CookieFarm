# Core Package Analysis

> **Package:** `server/core`
> **Files:** `config.go`, `flag_loop.go`, `flag_ttl.go`
> **Language:** Go

---

## 1. Component Inventory

### 1.1 `Runner` (struct — `config.go`)

The central orchestrator of the `core` package. It holds a reference to the database store and acts as the entry point for starting the flag processing background loops.

| Field   | Type               | Description                                      |
|---------|--------------------|--------------------------------------------------|
| `store` | `*database.Store`  | Reference to the database store (private field)  |

---

### 1.2 `NewRunner` (constructor — `config.go`)

**Signature:**
```go
func NewRunner(s *database.Store) *Runner
```

| | Type | Description |
|---|---|---|
| **Input** | `*database.Store` | An initialized database store |
| **Output** | `*Runner` | A new Runner instance |

**What it does:** Simple constructor. Allocates a `Runner` and wires in the store.

---

### 1.3 `Runner.Run` (method — `config.go`)

**Signature:**
```go
func (s *Runner) Run()
```

| | Type | Description |
|---|---|---|
| **Input** | — | None (receiver only) |
| **Output** | — | None (side effects only) |

**What it does:**
1. Creates a cancellable `context.Context`.
2. If a previous context cancel function (`shutdownCancel`) exists, calls it to stop the previous loops — this makes `Run()` safely re-entrant (useful when config is hot-reloaded via the web UI).
3. Stores the new cancel function in the package-level `shutdownCancel`.
4. Spawns `StartFlagProcessingLoop` as a goroutine.
5. If `FlagTTL != 0` in config, spawns `ValidateFlagTTL` as a goroutine.

---

### 1.4 `LoadConfigAndRun` (function — `config.go`)

**Signature:**
```go
func LoadConfigAndRun(path string, store *database.Store) error
```

| | Type | Description |
|---|---|---|
| **Input** | `path string` | Filesystem path to the YAML config file |
| **Input** | `store *database.Store` | An initialized database store |
| **Output** | `error` | `nil` on success, error on file/parse failure |

**What it does:**
1. Checks if the file at `path` exists.
2. Reads and YAML-unmarshals the file into `config.SharedConfig` (the global config singleton).
3. Sets `config.SharedConfig.Configured = true`.
4. Constructs a `Runner` and calls `Run()` on it.

---

### 1.5 `shutdownCancel` (package-level variable — `flag_loop.go`)

```go
var shutdownCancel context.CancelFunc
```

A package-level variable holding the cancel function for the currently active context. Allows `Run()` to tear down existing goroutines before spawning new ones. **This is the only state that lives at package level.**

---

### 1.6 `Runner.StartFlagProcessingLoop` (method — `flag_loop.go`)

**Signature:**
```go
func (s *Runner) StartFlagProcessingLoop(ctx context.Context)
```

| | Type | Description |
|---|---|---|
| **Input** | `ctx context.Context` | Cancellable context; cancelling it stops the loop |
| **Output** | — | None (blocking loop, runs until `ctx` is cancelled) |

**What it does:**
1. Creates a `time.Ticker` with interval `SubmitFlagCheckerTime` seconds.
2. Loads the submission protocol plugin via `protocols.LoadProtocol(...)` and stores the resulting function in `config.Submit`.
3. On each tick:
   - Fetches up to `MaxFlagBatchSize` unsubmitted flag codes from the DB.
   - If none found, skips.
   - Calls `config.Submit(url, token, flags)` to submit to the external flag checker.
   - Calls `s.UpdateFlags(responses)` to persist the results.
4. On `ctx.Done()`: stops and returns.

---

### 1.7 `Runner.UpdateFlags` (method — `flag_loop.go`)

**Signature:**
```go
func (s *Runner) UpdateFlags(flags []protocols.ResponseProtocol)
```

| | Type | Description |
|---|---|---|
| **Input** | `[]protocols.ResponseProtocol` | Slice of responses from the external flag checker |
| **Output** | — | None (side effects: DB updates + log summary) |

**What it does:**
1. Counts responses by status: `ACCEPTED`, `DENIED`, `ERROR`.
2. **Filters out** any response whose status is not one of those three known values (guards against unexpected protocol responses).
3. For each valid response, calls `store.Queries.UpdateFlagStatusByCode(...)` to persist the new status.
4. Logs a summary of accepted/denied/errored/total counts.

---

### 1.8 `Runner.ValidateFlagTTL` (method — `flag_ttl.go`)

**Signature:**
```go
func (s *Runner) ValidateFlagTTL(ctx context.Context, flagTTL uint64, tickTime int)
```

| | Type | Description |
|---|---|---|
| **Input** | `ctx context.Context` | Cancellable context; cancelling it stops the loop |
| **Input** | `flagTTL uint64` | Time-to-live expressed in game ticks |
| **Input** | `tickTime int` | Duration of one game tick in seconds |
| **Output** | — | None (blocking loop, runs until `ctx` is cancelled) |

**What it does:**
1. Computes the interval as `flagTTL * tickTime` seconds.
2. On each tick, calls `store.Queries.DeleteFlagByTTL(ctx, totalSecond)` to remove flags older than the TTL window.
3. Logs how many rows were deleted.
4. On `ctx.Done()`: stops and returns.

---

## 2. Dependency Graph

```
core
├-- server/config           (reads SharedConfig: SubmitFlagCheckerTime, MaxFlagBatchSize,
│                            URLFlagChecker, TeamToken, Protocol, FlagTTL, TickTime,
│                            Configured; writes config.Submit)
├-- server/database
│   ├-- Store               (holds *Queries)
│   └-- Queries
│       ├-- GetUnsubmittedFlagCodes()
│       ├-- UpdateFlagStatusByCode()
│       └-- DeleteFlagByTTL()
│       └-- MapFromResponseProtocolToParamsToUpdate()  [mapper.go]
├-- protocols (pkg)
│   ├-- LoadProtocol()      (loads .so plugin at runtime)
│   └-- ResponseProtocol    (Status, Flag, Msg)
├-- models (pkg)
│   └-- StatusAccepted/StatusDenied/StatusError constants
└-- logger (pkg)
    └-- Log                 (zerolog structured logger)
```

### Per-method dependency view

| Method | External deps |
|---|---|
| `NewRunner` | `database.Store` |
| `Run` | `config.SharedConfig`, `StartFlagProcessingLoop`, `ValidateFlagTTL`, `logger` |
| `LoadConfigAndRun` | `os`, `yaml.v3`, `config.SharedConfig`, `NewRunner`, `logger` |
| `StartFlagProcessingLoop` | `config.SharedConfig`, `config.Submit`, `protocols.LoadProtocol`, `database.Queries`, `logger` |
| `UpdateFlags` | `models` (status constants), `database.Queries`, `database.MapFromResponseProtocolToParamsToUpdate`, `logger` |
| `ValidateFlagTTL` | `database.Queries`, `logger` |

---

## 3. Consistency & Design Observations

### 3.1 What makes sense ✅

- The package has a single, clear responsibility: **orchestrate the background flag lifecycle** (submit to checker + purge expired flags).
- `Runner` is a clean value-object: it only carries the store and has no hidden global state in itself.
- `Run()` being re-entrant via `shutdownCancel` is correct and necessary for hot-reload of config.
- Separating `StartFlagProcessingLoop` and `ValidateFlagTTL` into distinct goroutines and files is logical.
- The invalid-status filter in `UpdateFlags` is a good defensive guard.

---

## 4. Impact Analysis (What breaks if you touch X)

| Change | Impact |
|---|---|
| Move `shutdownCancel` into `Runner` | Only `core` package internals; no external callers affected |
| Move file-loading out of `LoadConfigAndRun` | Callers: `server/cmd/root.go` calls `core.LoadConfigAndRun` — signature change needed there |
| Move `submitFunc` off `config.Submit` into `Runner` | `server/api/handler_api.go` calls `config.Submit` directly in `HandlePostFlag` and `HandlePostFlagsStandalone` — those handlers would need to receive the submit function via the handler or via `Runner` |
| Change `MaxFlagBatchSize` from `uint` to `int64` | `models.ConfigServer` in `pkg/models` — JSON/YAML serialization is unchanged; `handler_api.go` and any test that creates the struct needs updating |

---

## 5. Test Design (Category Partitioning)

### 5.1 `Runner.Run()`

| Category | Input | Expected Output |
|---|---|---|
| First call, TTL disabled | `store` valid, `FlagTTL == 0` | One goroutine spawned (`StartFlagProcessingLoop`), no TTL goroutine |
| First call, TTL enabled | `store` valid, `FlagTTL > 0` | Two goroutines spawned |
| Re-entrant call | Call `Run()` twice | Previous context cancelled, new goroutines replace old ones |
| Nil store | `store == nil` | No panic (goroutines fail gracefully when querying) |

### 5.2 `LoadConfig()`

| Category | Input | Expected Output |
|---|---|---|
| Valid path, valid YAML | Correct file | `nil` error, `config.SharedConfig.Configured == true`, `Run()` called |
| Non-existent path | Path to missing file | `error` returned, config unchanged |
| Existent path, invalid YAML | Malformed YAML | `error` returned, config unchanged |
| Empty path string | `""` | `error` returned (file does not exist) |

### 5.3 `Runner.UpdateFlags()`

| Category | Input | Expected Output |
|---|---|---|
| Empty slice | `[]ResponseProtocol{}` | No DB calls, log shows 0/0/0 |
| All accepted | All `Status == "ACCEPTED"` | All rows updated, accepted count correct |
| All denied | All `Status == "DENIED"` | All rows updated, denied count correct |
| Mixed statuses | Mix of ACCEPTED/DENIED/ERROR | Each counted and updated correctly |
| Unknown status | `Status == "SOMETHING_WEIRD"` | Entry filtered out, no DB call for it |
| DB update failure | Store returns error on update | Error logged, loop continues for remaining flags |

### 5.4 `Runner.ValidateFlagTTL()`

| Category | Input | Expected Output |
|---|---|---|
| Normal operation | `flagTTL=1`, `tickTime=60` | Ticker fires every 60s, `DeleteFlagByTTL` called |
| Zero affected rows | DB returns `(0, nil)` | Debug log, no error |
| Positive affected rows | DB returns `(5, nil)` | Info log with count |
| DB error | DB returns `(0, err)` | Error logged, loop continues |
| Context cancelled | `ctx` cancelled before tick | Loop exits cleanly, no deletion |

### 5.5 `Runner.StartFlagProcessingLoop()`

| Category | Input | Expected Output |
|---|---|---|
| Protocol load failure | `Protocol` name points to missing `.so` | Error logged, function returns immediately |
| No unsubmitted flags | DB returns empty slice | Debug "no flags" log, no submit call |
| Flags available, submit OK | DB returns flags, submit succeeds | `UpdateFlags` called with responses |
| Flags available, submit error | Submit returns error | Error logged, loop continues on next tick |
| DB fetch error | `GetUnsubmittedFlagCodes` returns error | Error logged, loop continues on next tick |
| Context cancelled mid-loop | `ctx` cancelled | Loop exits cleanly |
