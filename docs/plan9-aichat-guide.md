# Plan 9 aichat Quick Reference Guide
## Architecture, Features, and Usage at a Glance

---

## System Overview Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                  Plan 9 Distributed Network                   │
├──────────────────────────────────────────────────────────────┤
│                                                                │
│  TERMINAL                          CPU SERVER                  │
│  ┌─────────────────────┐          ┌──────────────────────┐    │
│  │   rc shell          │          │  aichat 9P Server    │    │
│  │ ┌─────────────────┐ │  9P      │ ┌────────────────┐   │    │
│  │ │ ai "query"      ├─┼─────────→│ │ Handle attach  │   │    │
│  │ │ aichat (REPL)   │ │  Protocol │ │ walk           │   │    │
│  │ │ aireview file   │ │          │ │ read/write     │   │    │
│  │ │ aigen code      │ │          │ └────────────────┘   │    │
│  │ └─────────────────┘ │          │                       │    │
│  │                     │          │ /ai/sessions/        │    │
│  │ /ai/                │          │ /ai/features/        │    │
│  │ (mounted)           │          │ /ai/ctl              │    │
│  └─────────────────────┘          └──────────────────────┘    │
│                                            ↓                    │
│                                  ┌─────────────────────┐       │
│                                  │ aichat CLI (Rust)   │       │
│                                  │ + LLM API Clients   │       │
│                                  └─────────────────────┘       │
│                                                                │
│  SECURITY BOUNDARY: Each process has isolated /ai/sessions/   │
│  AUTHENTICATION: factotum integration                          │
│  AUTHORIZATION: Per-process Qid capabilities                  │
│                                                                │
└──────────────────────────────────────────────────────────────┘
```

---

## Feature Card Display Examples

### Card 1: Shell Command Assistant
```
╔═══════════════════════════════════════════════════════════════╗
║ Shell Command Assistant                     [shell-assist-01] ║
╠═══════════════════════════════════════════════════════════════╣
║                                                               ║
║ Parse natural language descriptions into syntactically       ║
║ correct rc shell commands with Plan 9 idiom compliance.     ║
║                                                               ║
║ ✓ Syntax validation              ✓ Error explanation        ║
║ ✓ Plan 9 idiom compliance        ✓ Alternative suggestions  ║
║                                                               ║
║ Model: claude-3.5-sonnet | Tokens: 512 | Temp: 0.3           ║
║                                                               ║
║ Status: Ready  |  Sessions: 3  |  Avg Response: 245ms        ║
║                                                               ║
╠═══════════════════════════════════════════════════════════════╣
║ [Use Feature] [View Examples] [Adjust Settings] [View Docs]   ║
╚═══════════════════════════════════════════════════════════════╝
```

### Card 2: Code Review
```
╔═══════════════════════════════════════════════════════════════╗
║ rc Script Reviewer                           [code-review-01] ║
╠═══════════════════════════════════════════════════════════════╣
║                                                               ║
║ Analyze rc shell scripts for correctness, performance, and   ║
║ adherence to Plan 9 best practices.                           ║
║                                                               ║
║ ✓ Syntax checking               ✓ Performance optimization   ║
║ ✓ Idiom verification            ✓ Security analysis          ║
║                                                               ║
║ Model: claude-3.5-sonnet | Tokens: 1024 | Temp: 0.2          ║
║                                                               ║
║ Status: Ready  |  Scripts reviewed: 47  |  Avg: 312ms        ║
║                                                               ║
╠═══════════════════════════════════════════════════════════════╣
║ [Start Review] [Batch Process] [View Reports] [Settings]      ║
╚═══════════════════════════════════════════════════════════════╝
```

---

## File Hierarchy (Actual Filesystem)

```
/ai/
│
├── ctl                         ← Read: server status
│                                 Write: create new session
│
├── models                      ← Available LLM models
│   ├── claude-3.5-sonnet
│   ├── gpt-4-turbo
│   └── gpt-4o
│
├── config                      ← Global configuration
│   ├── default_model
│   ├── token_limit
│   └── temperature
│
├── features/                   ← Public feature catalog
│   │
│   ├── shell-assist/
│   │   ├── meta               ← Feature metadata (JSON)
│   │   ├── prompt             ← System prompt
│   │   ├── capabilities       ← Feature list
│   │   └── examples           ← Example exchanges
│   │
│   ├── code-review/
│   │   ├── meta
│   │   ├── prompt
│   │   ├── capabilities
│   │   └── examples
│   │
│   └── rc-explain/
│       └── ...
│
└── sessions/                   ← Per-process sessions
    │                              (kernel isolated)
    │
    ├── s12345/                 ← Session for PID 12345
    │   ├── meta                ← Metadata
    │   │   - id: s12345
    │   │   - user: alice
    │   │   - pid: 12345
    │   │   - created: 1737193530
    │   │
    │   ├── state               ← Current configuration
    │   │   - model: claude
    │   │   - tokens: 512
    │   │
    │   ├── context             ← Accumulated context
    │   │   - pwd: /home/alice/src
    │   │   - user: alice
    │   │   - shell: /bin/rc
    │   │
    │   ├── history             ← Message log
    │   │   Q: find .c files > 100KB
    │   │   A: find . -name "*.c" -size +100k
    │   │   Q: explain that
    │   │   A: This finds all .c files...
    │   │
    │   └── ctl                 ← Control (write queries)
    │                              (read responses)
    │
    ├── s54321/                 ← Another session
    │   └── ...
    │
    └── s99999/
        └── ...
```

---

## Shell Command Reference

### Basic Usage
```rc
# Simple query
ai 'find all .go files'

# Multi-word query
ai find all .go files modified in last 7 days

# REPL mode
aichat          # Interactive conversation

# Explain command
aiexplain 'ls -la'

# Review script
aireview myscript.rc

# Generate code
aigen 'function to read config file'

# Find command
aifind 'extract tar.gz files'

# Batch process
aibatch 'optimize this script' *.rc
```

### Advanced Usage
```rc
# Pipeline integration
{
    echo 'convert bytes to megabytes'
    echo '12582912'
} | ai

# Multi-turn context
aichat
  ai> explain Plan 9 mounting
  ai> how does it differ from Linux?
  ai> show me an example
  ai> exit

# Feature card usage
aifeature shell-assist 'best way to find large files'

# Check service status
cat /ai/ctl

# View available features
ls /ai/features

# View session info
cat /ai/sessions/s$pid/meta

# View history
cat /ai/sessions/s$pid/history

# Access feature metadata
cat /ai/features/shell-assist/prompt
```

---

## Session Lifecycle

```
User starts rc shell
    │
    ├─→ import /ai (mount aichat service)
    │
    ├─→ First ai command
    │   ├─→ Server auto-creates /ai/sessions/s{pid}/
    │   ├─→ Initializes meta, context, history files
    │   └─→ Binds session to process
    │
    ├─→ Subsequent ai commands
    │   ├─→ Read existing /ai/sessions/s{pid}/
    │   ├─→ Update context if needed
    │   ├─→ Write query to ctl
    │   ├─→ Read response from history
    │   └─→ Add to accumulated context
    │
    └─→ Shell exit (cleanup)
        ├─→ Archive /ai/sessions/s{pid}/history
        ├─→ Close session
        └─→ Session directory remains for audit trail
```

---

## Architecture Components

| Component | Type | Language | Role |
|-----------|------|----------|------|
| **aichat** | 9P Server | C | Handles filesystem protocol, session mgmt |
| **aichat (CLI)** | Backend | Rust | Actual LLM interaction, model selection |
| **aichat.rc** | Shell Lib | rc | User-facing functions, REPL, helpers |
| **feature JSONs** | Metadata | JSON | Card definitions, prompts, examples |
| **factotum** | Auth | Existing | Authentication provider |
| **Kernel** | OS | C | Per-process namespace isolation |

---

## Key Operations

### Operation 1: Create Query
```
User enters: ai 'list all README files'
    ↓
Function ai() called
    ├─ Ensure mount: aichat_ensure_mount
    ├─ Get session: aichat_getsession → s{pid}
    ├─ Write query: echo "list all README files" > /ai/sessions/s{pid}/ctl
    └─ Wait for response
    ↓
Server (aichat.c)
    ├─ Detect write to Qsessctl
    ├─ Extract message
    ├─ Invoke aichat CLI: echo "list all README files" | aichat -m claude...
    └─ Capture response
    ↓
Server adds to history
    ├─ addhistory(session, response)
    ├─ Append to circular buffer
    └─ Update mtime
    ↓
User reads result
    ├─ Function returns: cat /ai/sessions/s{pid}/history | tail -1
    ├─ Displays: find . -name 'README*'
    └─ Command ready to execute
```

### Operation 2: Multi-Turn Conversation
```
Session 1: "explain Plan 9 mounting"
    ├─ Update context: {pwd, user, shell}
    ├─ Query sent
    └─ Response: "Plan 9 mounting is..."
    
Session 2: "how does it differ from Linux?"
    ├─ Context available from previous
    ├─ Server can factor in previous context
    ├─ Query references prior conversation
    └─ Response: "Unlike Linux mount, Plan 9..."
    
Session 3: "show me an example"
    ├─ Full conversation history available
    ├─ LLM can provide coherent examples
    └─ Response: "rc; mount -c /mnt/cpu/..."
```

---

## Security Model

### Authentication
```
┌─────────────┐
│ factotum    │  ← Centralized key/auth management
└──────┬──────┘
       │ Provides credentials
       │
       ▼
┌─────────────────────────────┐
│ User starts rc shell        │
├─────────────────────────────┤
│ attach to /ai service       │
│  - factotum validates       │
│  - session created          │
│  - Qid capability issued    │
└──────┬──────────────────────┘
       │
       ▼
   Can read/write own session
   Can read public features
   Cannot access other sessions
```

### Authorization
```
File Access Control (Qid-based)

/ai/ctl
  ├─ Owner: aichat server
  └─ Read: public (R-)
  └─ Write: owner only (-W)

/ai/features/*
  ├─ Owner: aichat server
  └─ Read: public (R-)
  └─ Write: none

/ai/sessions/s{pid}/
  ├─ Owner: process owner
  └─ Read: owner only (R-)
  └─ Write: owner only (-W)
  └─ Enforce: kernel (per-process ns)

/ai/sessions/{other}/
  ├─ Cannot walk
  ├─ Cannot read
  └─ Cannot write
```

---

## Deployment Checklist

- [ ] **Build**: Compile aichat.c with 9P headers
- [ ] **Install**: Copy binary to /bin/aichat
- [ ] **Library**: Install aichat.rc to /lib/aichat/
- [ ] **Features**: Deploy JSON files to /lib/aichat/features/
- [ ] **Config**: Set up /etc/aichat/config
- [ ] **Mount**: Add to /etc/profile: mount -c /mnt/cpu/srv/aichat /ai
- [ ] **Shell Integration**: Source aichat.rc in ~/.rcrc
- [ ] **Test**: Run `ai 'test query'`
- [ ] **Verify**: Check `/ai/sessions/s$pid/history`
- [ ] **Monitor**: Watch `/ai/ctl` for status

---

## Troubleshooting

### Service not mounted
```rc
test -d /ai || mount -c /mnt/cpu/srv/aichat /ai
```

### Session not found
```rc
# Create new session
aichat_ensure_mount
# Session auto-created on first use
ai 'test'
```

### History not updating
```rc
# Check session ctl
ls -la /ai/sessions/s$pid/
# Verify server is running
cat /ai/ctl
```

### Slow responses
```rc
# Check model load
cat /ai/ctl | grep model
# Try reducing token limit
echo 'tokens: 256' > /ai/sessions/s$pid/state
```

---

## Performance Characteristics

| Metric | Value | Notes |
|--------|-------|-------|
| Session creation | <100ms | Auto-created on first access |
| Query latency | 500ms-2s | Depends on LLM response time |
| History buffer | 256 entries | Circular, memory efficient |
| Max concurrent sessions | 64 | Configurable |
| Protocol overhead | <10ms | 9P is lightweight |
| Context size | 8KB | Per session |

---

## Integration with aichat (sigoden)

The plan9-aichat implementation wraps the existing aichat CLI tool:

```
aichat (sigoden's Rust tool)
  ├─ Models: claude-3.5-sonnet, gpt-4*, gpt-4o
  ├─ Features: REPL, shell assistant, RAG, tools
  └─ Output: Structured text responses

Wrapped by: aichat.c (9P server)
  ├─ Protocol: 9P file access
  ├─ Sessions: Per-process isolation
  ├─ Storage: Filesystem-based history
  └─ Distribution: Works across Plan 9 network
```

This extends aichat from a CLI tool to a **distributed service** while maintaining compatibility with its existing capabilities.

---

## References

- **Plan 9 Documentation**: /sys/doc/
- **9P Protocol**: 9p(5) manual page
- **rc Shell**: rc(1) manual page  
- **factotum**: factotum(4) manual page
- **aichat**: https://github.com/sigoden/aichat
- **Architecture Pattern**: Everything is a file

---

*Designed for Johannesburg, South Africa aichat enthusiasts*
*Friday, January 18, 2026 | 07:11 AM SAST*