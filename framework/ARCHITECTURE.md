# 9cog Framework Architecture

## Component Integration Map

```
                         ┌─────────────────┐
                         │   User (airc)    │
                         │   rc shell with  │
                         │   AI integration │
                         └────────┬─────────┘
                                  │ file I/O
                         ┌────────▼─────────┐
                         │  AI Filesystem   │
                         │   (9fs9rc)       │
                         │  /ai/sessions/   │
                         │  /ai/models/     │
                         │  /ai/knowledge/  │
                         │  /ai/topology/   │
                         └───┬────┬────┬────┘
                    ┌────────┘    │    └────────┐
                    ▼            ▼             ▼
           ┌──────────────┐ ┌─────────┐ ┌──────────────┐
           │  AI Engine   │ │Knowledge│ │  Topology    │
           │  (aichat)    │ │  Graph  │ │  (120c)      │
           │  20+ LLM     │ │(cogpwsh)│ │  120-cell    │
           │  providers   │ │OpenCog  │ │  Phoenix     │
           └──────────────┘ └─────────┘ └──────────────┘
                    │
           ┌────────▼─────────┐
           │   9P Protocol    │
           │ go9p │ diod │    │
           │ lua9p│ plan9│    │
           │ TCP/RDMA/FD      │
           └──────────────────┘
```

See the [full architecture document](https://github.com/9cog/9fs9rc/blob/claude/consolidate-repos-framework-Uy8EC/framework/README.md) for details on data flow, repository roles, and integration patterns.
