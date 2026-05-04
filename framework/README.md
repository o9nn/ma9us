# 9cog: Cognitive 9P Framework

A unified framework that integrates Plan 9 distributed computing, AI services,
and cognitive architecture into a single cohesive system.

## Vision

**Everything is a file. Everything is intelligent.**

9cog extends Plan 9's "everything is a file" philosophy to AI and cognitive
computing. AI services, knowledge graphs, and organizational models are all
exposed as 9P file servers, composable through the rc shell and accessible
from any 9P client.

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                     User Interface Layer                      │
│  airc (rc shell) ─── REPL ─── CLI ─── Web Playground        │
├──────────────────────────────────────────────────────────────┤
│                     AI Services Layer                         │
│  aichat engine: 20+ LLM providers, RAG, agents, functions   │
├──────────────────────────────────────────────────────────────┤
│                     Cognitive Layer                           │
│  cogpwsh: Knowledge graph, pattern matching, truth values    │
│  120c: Organizational topology, Phoenix Engine lifecycle     │
├──────────────────────────────────────────────────────────────┤
│                     Filesystem Layer                          │
│  /ai/sessions/  /ai/models/  /ai/features/  /ai/knowledge/  │
│  Synthetic 9P files exposing all services as file operations │
├──────────────────────────────────────────────────────────────┤
│                     9P Protocol Layer                         │
│  go9p (Go) ── diod (C) ── lua9p (Lua) ── plan9/lib9p (C)   │
│  TCP, RDMA, Unix socket, FD transports                       │
└──────────────────────────────────────────────────────────────┘
```

## Repository Map

| Repository | Role | Language | Layer |
|------------|------|----------|-------|
| [9fs9rc](https://github.com/9cog/9fs9rc) | **Integration hub** — framework core, AI filesystem design | Go/C | Filesystem |
| [go9p](https://github.com/9-p9/go9p) | 9P client/server library + AI filesystem server | Go | Protocol |
| [diod](https://github.com/9-p9/diod) | Production 9P2000.L server with multi-transport | C | Protocol |
| [lua9p](https://github.com/9-p9/lua9p) | Lightweight 9P client for scripting | Lua | Protocol |
| [plan9](https://github.com/9-p9/plan9) | Plan 9 4th Ed — reference implementation, lib9p | C | Protocol |
| [aichat](https://github.com/9cog/aichat) | Multi-LLM engine with RAG, agents, function calling | Rust | AI Services |
| [cogpwsh](https://github.com/9cog/cogpwsh) | OpenCog knowledge graph + GitHub API | PowerShell | Cognitive |
| [120c](https://github.com/9cog/120c) | Organizational topology modeling + visualization | Python | Cognitive |
| [airc](https://github.com/9cog/airc) | rc shell with AI tool integration | C | Interface |

## Quick Start

### 1. Start the AI filesystem server

```bash
go run ./cmd/aifs -listen :5641 -provider openai -model gpt-4
```

### 2. Mount and interact from rc shell

```rc
9mount localhost:5641 /ai
echo 'explain quicksort' > /ai/sessions/new/ctl
cat /ai/sessions/1/history
```

### 3. Query the knowledge graph

```rc
cat /ai/knowledge/concepts/quicksort
echo 'pattern InheritanceLink($x, Algorithm)' > /ai/knowledge/query
cat /ai/knowledge/results
```

### 4. Use from any language via 9P

```lua
local p9 = require '9p'
local conn = p9.newconn(read, write)
p9.attach(conn, "user", "/ai")
local fid = p9.walk(conn, conn.rootfid, nil, "sessions/new/ctl")
p9.open(conn, fid, p9.OWRITE)
p9.write(conn, fid, 0, "what is 9P?")
```

## Core Concepts

### AI-as-Filesystem

```
/ai/
├── features/           # Queryable AI capabilities
├── sessions/           # Per-process conversation state
│   └── {id}/
│       ├── ctl         # Control: write commands, read status
│       ├── history     # Full conversation log
│       ├── context     # Session context (pwd, shell, etc.)
│       └── meta        # Session metadata (model, tokens, etc.)
├── models/             # Available LLM backends
├── knowledge/          # Cognitive layer
├── topology/           # Organizational modeling (120-cell)
└── config              # Framework configuration
```

## Design Principles

1. **Everything is a file** — All services accessible via 9P file operations
2. **Composability** — Shell pipes connect AI, knowledge, and visualization
3. **Per-process namespaces** — Each shell gets isolated AI sessions
4. **Protocol-first** — 9P2000 as the universal integration protocol
5. **Language-agnostic** — Go, C, Lua, Rust, Python, PowerShell all participate
6. **Distributed by default** — Services can run on different machines
7. **Cognitive awareness** — Knowledge graph tracks system state and relationships

## License

Individual components retain their original licenses. Framework integration
code is MIT licensed.
