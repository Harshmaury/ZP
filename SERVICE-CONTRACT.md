// @zp-project: zp
// @zp-path: SERVICE-CONTRACT.md
# SERVICE-CONTRACT.md — ZP
# @version: 2.0.0
# @updated: 2026-03-25

**Type:** CLI tool · **Module:** `github.com/Harshmaury/ZP` · **Domain:** Tool

---

## Code

```
cmd/zp/main.go              CLI entry, command dispatch
internal/manifest/manifest.go  nexus.yaml parse -- ProjectManifest
internal/registry/registry.go  workspace scan -- project discovery
internal/pack/zipper.go       file collection + ZIP write
internal/pack/filter.go       -H / -go / -yaml / -api / -core / -pkg / -store / -config
internal/pack/dev.go          sandbox in /tmp/zp-dev/
internal/gate/arbiter.go      VerifyPackaging gate -- blocks on violations
internal/config/config.go     ZP_DROP_DIR, ZP_WORKSPACE env vars
```

---

## Contract

**Commands:**
```
zp                    package current project
zp <id>               package by ID
zp <id> <id> ...      multiple -- combined ZIP
zp all                all workspace projects
zp list | ls          list discovered projects
zp status | st        projects + last ZIP timestamp
zp dev <id>           sandbox in /tmp/zp-dev/
zp version / help
```

**ZIP naming:** `<project>-<filter>-<YYYYMMDD>-<HHMM>.zip`

**Always included:** `go.mod`, `go.sum`, `nexus.yaml`, `.zpignore`, `WORKFLOW-SESSION.md`

**Always excluded:** directories starting with `_` or `.`

**Arbiter gate:** `VerifyPackaging(dir)` runs before any ZIP is written. `--skip-enforce` bypasses and emits `SYSTEM_ALERT` to Nexus.

---

## Control

Single-threaded CLI. No platform API calls during normal operation. Registry scan is point-in-time.

---

## Context

No HTTP calls to any platform service. Read-and-package only.
