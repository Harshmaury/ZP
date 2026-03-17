# ADR-019 — zp: Developer Packaging Tool

**Status:** Accepted
**Date:** 2026-03-17
**Author:** Harsh Maury
**Scope:** Developer tooling — zp binary
**Location:** ~/workspace/projects/tools/zp

---

## Context

The platform has 8 services across multiple repositories. Every
development session requires packaging specific files into ZIPs and
delivering them to the engx-drop folder. This is currently done
manually — selecting files, naming ZIPs, copying to the right path.

As the platform grows, this friction compounds:
- Wrong files included/excluded
- Inconsistent ZIP naming
- No filter by layer (handlers, config, API, core)
- No way to package cross-service contract sets

zp is the platform's first developer tool — a Go binary that reads
nexus.yaml, understands the project structure, and produces
consistently named, correctly filtered ZIPs in one command.

---

## Decision

### 1. zp is a standalone Go binary

Not a bash script. Go gives us:
- nexus.yaml parsing (project identity)
- Glob pattern matching for filters
- Consistent cross-platform path handling
- Clean error messages

### 2. Core commands

```
zp                     # package current project (reads nexus.yaml)
zp nexus               # package named project by id
zp atlas forge         # package multiple projects in one run
zp all                 # package every registered platform project

zp -H                  # handlers only (internal/api/handler/)
zp -go                 # Go source files only
zp -yaml               # YAML/config files only
zp -api                # full API layer (handler + server + middleware)
zp -core               # core logic (non-API, non-cmd)

zp dev <project>       # create isolated dev sandbox in /tmp/zp-dev/
zp help                # clean, readable help output
```

### 3. ZIP naming convention (enforced)

```
<project>-<filter>-<YYYYMMDD>-<HHMM>.zip
nexus-full-20260317-2144.zip
forge-handlers-20260317-2144.zip
atlas-forge-full-20260317-2144.zip   (multi-project)
```

### 4. Output destination

Default: ~/Downloads/nexus-drop/ (engx-drop)
Override: ZP_DROP_DIR env var or --out flag

### 5. nexus.yaml awareness

zp reads nexus.yaml in the target project dir to:
- Confirm project ID for ZIP naming
- Exclude files listed in .zpignore (if present)
- Auto-detect project root

### 6. .zpignore support

Projects can define .zpignore (gitignore syntax) to exclude:
- vendor/, node_modules/, .git/
- *_test.go (optional)
- *.pb.go (generated files)

Default excludes always applied:
- .git/, vendor/, node_modules/
- *.exe, *.dll, *.so
- *_test.go

### 7. dev isolation mode

zp dev <project> creates:
```
/tmp/zp-dev/<project>-<timestamp>/
  <project>/          ← full project copy
  contracts/          ← nexus.yaml files from all depends_on projects
  README.md           ← what's here and why
```

This gives a clean working context without touching the live workspace.

---

## Compliance

- Does not modify any platform service
- Does not communicate with any platform API
- Pure filesystem tool — no ADR constraints apply
- Replaces manual ZIP workflow entirely

---

## Next ADR

ADR-020 — devtest contract validation tool (if needed).
