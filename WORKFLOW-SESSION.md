# WORKFLOW-SESSION.md
# Session: ZP-fix-manifest-loadfromid
# Date: 2026-03-20

## What changed — remove LoadFromID dead code

LoadFromID in internal/manifest/manifest.go contained hardcoded search
paths that assumed the old workspace layout (projects/apps/, projects/tools/).
The actual workspace is projects/engx/services/. Any zp command that missed
the registry scan would fall through to LoadFromID and fail to find the project.

LoadFromID was marked as a backwards-compat fallback, but the registry scan
(registry.Scan) already performs a depth-4 walk across the entire workspace
and finds any project with a nexus.yaml regardless of directory structure.
The fallback was dead code that silently broke under workspace reorganisation.

Fix: delete LoadFromID entirely. resolveProject in cmd/zp/main.go now returns
a clear error if registry.Find returns nil — no silent fallback, no stale paths.

## Files changed

- `internal/manifest/manifest.go`  — deleted LoadFromID function
- `cmd/zp/main.go`                 — removed LoadFromID call from resolveProject

## Apply

```bash
cd ~/workspace/projects/engx/services/zp && \
unzip -o /mnt/c/Users/harsh/Downloads/engx-drop/zp-fix-manifest-loadfromid-20260320.zip -d . && \
go build ./...
```

## Verify

```bash
go build ./...

# From anywhere in the workspace — should find nexus by registry scan:
zp list
zp nexus

# Should give clear error for unknown project (not a path-resolution failure):
zp doesnotexist
# Expected: project "doesnotexist" not found — run 'zp list' to see available projects
```

## Commit

```bash
git add \
  internal/manifest/manifest.go \
  cmd/zp/main.go \
  WORKFLOW-SESSION.md && \
git commit -m "fix(manifest): remove LoadFromID — stale hardcoded paths, registry scan is sufficient" && \
git push origin main
```
