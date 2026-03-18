# SERVICE-CONTRACT.md — zp

**Service:** zp
**Domain:** Tool (CLI)
**Port:** none
**ADRs:** ADR-019 (packaging tool)
**Version:** 2.0.0
**Updated:** 2026-03-18

---

## Role

Developer packaging tool. Reads `nexus.yaml`, discovers platform projects,
and produces consistently named ZIPs for delivery to the engx-drop folder.
Replaces all manual ZIP creation. Pure filesystem tool — no platform API calls.

---

## Inputs

- `nexus.yaml` in project roots — project ID and metadata
- Workspace filesystem — source files to package
- `ZP_DROP_DIR` env var — output directory override
- `ZP_WORKSPACE` env var — workspace root override
- CLI arguments — project IDs, filter flags, command

---

## Outputs

- ZIP files in drop directory: `<project>-<filter>-<YYYYMMDD>-<HHMM>.zip`
- Console output: project, filter, file count, output path
- Dev sandbox: `/tmp/zp-dev/<project>-<ts>/` with project copy + contracts

---

## Commands

```
zp                    package current project
zp <id>               package by ID
zp <id> <id> ...      package multiple → combined ZIP
zp all                package all workspace projects
zp list / ls          list all discovered projects
zp status / st        show projects + last ZIP timestamp
zp dev <id>           create isolated dev sandbox
zp version            print version
zp help               print help
```

## Filters

`-H` handlers · `-go` Go files · `-yaml` YAML · `-api` API layer ·
`-core` core logic · `-pkg` pkg/ · `-store` store/ · `-config` config + YAML

## Flags

`--out <dir>` · `--path <dir>`

---

## Dependencies

None. zp makes no HTTP calls to any platform service.
It reads only the local filesystem and `nexus.yaml` files.

---

## Guarantees

- ZIP naming convention is always enforced — no manual naming.
- `go.mod`, `go.sum`, `nexus.yaml`, `.zpignore` always included regardless of filter.
- Directories starting with `_` or `.` are always excluded.
- `zp all` uses dynamic registry scan — no hardcoded project list.
- Registry scan is a point-in-time snapshot — projects modified during
  `zp all` may be skipped with an error (expected, non-fatal).

## Non-Responsibilities

- zp does not communicate with Nexus, Atlas, Forge, or any observer.
- zp does not register projects, start services, or trigger workflows.
- zp does not modify any platform database.

## Data Authority

None. zp is a read-and-package tool. It produces ZIPs, not platform state.

## Concurrency Model

Single-threaded CLI. No concurrent operations.
