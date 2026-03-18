# WORKFLOW-SESSION.md
# Session: ZP-v2-advanced-packaging
# Date: 2026-03-17

## What changed — zp v2.0.0

Major upgrade. 10 gaps fixed from deep research on actual codebase.

## Changes vs v1.0.0

### New commands
- zp list / zp ls    — discover all nexus.yaml projects in workspace
- zp status / zp st  — list projects + last ZIP timestamp

### New flags
- --out <dir>        — override output dir per command
- --path <dir>       — package arbitrary dir without nexus.yaml

### New filters
- -pkg               — pkg/ layer only (was missing from v1)
- -store             — internal/store/ only
- -config            — internal/config/ + YAML files

### Bug fixes
- _backups/ dirs now excluded (dirs starting with _ skipped)
- go.mod + go.sum always included regardless of filter mode
- nexus.yaml always included in all filter modes
- runAll now uses dynamic registry scan (no hardcoded list)
- resolveProject uses registry scan first, falls back to path search
- WSL2: auto-detects Windows home for engx-drop default path

## Apply

cd ~/workspace/projects/tools/zp
unzip -o /mnt/c/Users/harsh/Downloads/engx-drop/zp-v2-20260317.zip -d .
go mod tidy && go build ./...
go install ./cmd/zp/ && cp ~/go/bin/zp ~/bin/zp

## Verify

zp version          # should show 2.0.0
zp help
zp list
zp status
cd ~/workspace/projects/apps/nexus && zp
zp nexus -H
zp all

## Commit

git add . && git commit -m "feat: zp v2.0.0 — list, status, --out, --path, dynamic discovery" && \
git tag v2.0.0 && git push origin main --tags
