# WORKFLOW-SESSION.md
# Session: ZP-phase1-packaging-tool
# Date: 2026-03-17

## What changed — zp v1.0.0 (ADR-019)

New developer packaging tool. Replaces manual ZIP workflow entirely.

## Setup

mkdir -p ~/workspace/projects/tools/zp
cd ~/workspace/projects/tools/zp
unzip -o /mnt/c/Users/harsh/Downloads/engx-drop/zp-tool-20260317.zip -d .
go mod tidy && go build ./...
go install ./cmd/zp/ && cp ~/go/bin/zp ~/bin/zp

## Verify

zp help
zp version

# Package current project (from any project root)
cd ~/workspace/projects/apps/nexus && zp
cd ~/workspace/projects/apps/forge && zp -H

# Package by ID
zp nexus
zp atlas forge -api
zp all

# Dev sandbox
zp dev forge

## Commit

git init && git add . && \
git commit -m "feat: zp developer packaging tool v1.0.0 (ADR-019)" && \
git tag v1.0.0
