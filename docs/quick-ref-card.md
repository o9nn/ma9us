# Plan 9 aichat - Visual Architecture Card
## One-Page Reference for System Design

---

## SYSTEM LAYERS

```
┌─────────────────────────────────────────────────────────────────┐
│ USER LAYER                                                       │
│ rc shell functions: ai, aichat, aiexplain, aireview, aigen     │
├─────────────────────────────────────────────────────────────────┤
│ FILESYSTEM LAYER                                                 │
│ /ai/sessions/{pid}/meta, context, history, ctl                 │
│ /ai/features/{name}/meta, prompt, examples                     │
├─────────────────────────────────────────────────────────────────┤
│ PROTOCOL LAYER                                                   │
│ 9P file system protocol (attach, walk, read, write)            │
├─────────────────────────────────────────────────────────────────┤
│ SERVER LAYER                                                     │
│ aichat.c: 9P server, session mgmt, history, backend            │
├─────────────────────────────────────────────────────────────────┤
│ BACKEND LAYER                                                    │
│ aichat CLI (Rust) + LLM APIs (Claude, GPT-4, etc)             │
└─────────────────────────────────────────────────────────────────┘
```

---

## SESSION ANATOMY

```
/ai/sessions/s12345/                      (per-process directory)
├── meta (read)                           (static metadata)
│   ├─ id: s12345
│   ├─ user: alice
│   └─ pid: 12345
│
├── context (read)                        (accumulated state)
│   ├─ pwd: /home/alice/src
│   ├─ user: alice
│   ├─ shell: /bin/rc
│   └─ bindings: [...]
│
├── history (read)                        (message log)
│   ├─ Q: find large files
│   ├─ A: find . -size +100m
│   ├─ Q: explain that
│   └─ A: This searches...
│
└── ctl (write→read)                      (query interface)
    Write query → Server processes → Returns via history
```

---

## FEATURE CARD ANATOMY

```
METADATA FILE: /ai/features/{name}/meta

{
  "id":        "shell-assist-001",
  "name":      "Shell Command Assistant",
  "desc":      "Natural language → rc commands",
  "caps":      ["syntax", "idioms", "safety"],
  "model":     "claude-3.5-sonnet",
  "temp":      0.3,
  "tokens":    512
}

VISUAL DISPLAY:

╔════════════════════════════════════╗
║ Feature Name          [id]         ║
╠════════════════════════════════════╣
║ Description here                   ║
║                                    ║
║ ✓ Capability 1    ✓ Capability 2  ║
║ ✓ Capability 3    ✓ Capability 4  ║
║                                    ║
║ Model: X | Temp: Y.Z | Tokens: N  ║
╠════════════════════════════════════╣
║ [Use] [Info] [Examples] [Settings] ║
╚════════════════════════════════════╝
```

---

## QUERY FLOW

```
User:
  ai 'find all .c files'
                 │
                 ▼
rc function ai():
  ✓ aichat_ensure_mount()     [/ai mounted?]
  ✓ aichat_getsession()        [PID → s{pid}]
  ✓ Write: echo "query" > ctl
  ✓ Read: cat history | tail -1
                 │
                 ▼
aichat.c (aichwrite):
  ✓ Parse Qsessctl write
  ✓ Extract message
  ✓ addhistory(session, msg)
  ✓ processquery(session, msg)
                 │
                 ▼
processquery():
  ✓ Build command: aichat -m claude...
  ✓ Execute via popen()
  ✓ Capture response
  ✓ addhistory(session, response)
                 │
                 ▼
User reads:
  cat /ai/sessions/s{pid}/history | tail -1
  Output: find . -name '*.c'
```

---

## SECURITY MODEL

```
AUTHENTICATION:
  factotum ← (credentials) → aichat server
                              ↓
                        Verify user

AUTHORIZATION (Qid-based):
  /ai/ctl
    Owner: aichat   | Public read | Owner write
  
  /ai/features/*
    Owner: aichat   | Public read | No write
  
  /ai/sessions/s{MY_PID}
    Owner: user     | Own read    | Own write
  
  /ai/sessions/s{OTHER_PID}
    Cannot walk     | Cannot read | Cannot write

ISOLATION (Kernel):
  Process A: only sees /ai/sessions/s{pidA}/
  Process B: only sees /ai/sessions/s{pidB}/
  Enforced by per-process namespaces
```

---

## COMMAND REFERENCE

```
BASIC:
  ai query                    Simple query
  aichat                      Interactive REPL

SPECIALIZED:
  aiexplain cmd               Explain command
  aireview file               Review script
  aigen description           Generate code
  aifind task                 Find command
  aibatch query files...      Batch process

INTERNAL:
  aichat_query text           Core query function
  aichat_getsession           Get current session
  aichat_ensure_mount         Verify /ai mounted
  aichat_feature_show name    Display feature card

FILES:
  cat /ai/ctl                 Server status
  cat /ai/models              List models
  ls /ai/features             List features
  cat /ai/sessions/s$pid/meta Session info
  cat /ai/sessions/s$pid/hist Message history
```

---

## WORKFLOW PATTERNS

```
PATTERN 1: SINGLE QUERY
┌─────────────────────────────────────────────┐
│ ai "find .c files"                          │
├─────────────────────────────────────────────┤
│ Output: find . -name "*.c"                  │
│ Status: ✓ Ready to execute                 │
└─────────────────────────────────────────────┘

PATTERN 2: MULTI-TURN CONTEXT
┌─────────────────────────────────────────────┐
│ ai "explain Plan 9 mounting"                │
├─────────────────────────────────────────────┤
│ Response: Plan 9 mounting is...             │
├─────────────────────────────────────────────┤
│ ai "how does it differ from Linux?"         │
├─────────────────────────────────────────────┤
│ Response: Unlike Linux mount...             │
└─────────────────────────────────────────────┘
  [Prior context automatically available]

PATTERN 3: FEATURE SELECTION
┌─────────────────────────────────────────────┐
│ aifeature shell-assist "large files"        │
├─────────────────────────────────────────────┤
│ [Display feature card]                      │
├─────────────────────────────────────────────┤
│ Execute query with feature's prompt         │
└─────────────────────────────────────────────┘

PATTERN 4: PIPELINE INTEGRATION
┌─────────────────────────────────────────────┐
│ {code} | ai "optimize this"                 │
├─────────────────────────────────────────────┤
│ Processes input + sends to aichat          │
└─────────────────────────────────────────────┘
```

---

## IMPLEMENTATION CHECKLIST

```
BEFORE DEPLOYMENT:
  ☐ Compile: 9c -o aichat aichat.c -l9p -lthread
  ☐ Install: cp aichat /bin/
  ☐ Library: cp aichat.rc /lib/aichat/
  ☐ Features: cp features/*.json /lib/aichat/features/
  ☐ Config: setup /etc/aichat/config

AT STARTUP:
  ☐ mount -c /mnt/cpu/srv/aichat /ai
  ☐ source /lib/aichat/aichat.rc

VERIFY:
  ☐ test -d /ai ✓
  ☐ ls /ai/features ✓
  ☐ ai "test" ✓
  ☐ cat /ai/sessions/s$pid/history ✓
```

---

## PERFORMANCE TARGETS

```
Operation                Time        Notes
─────────────────────────────────────────────
Session creation         <100ms      Auto on first use
Query submission         <10ms       9P write
Backend processing       500ms-2s    Depends on model
Response retrieval       <5ms        9P read
History buffer add       <1ms        Circular buffer
Context update           <2ms        File append
Multi-turn overhead      <5ms        Session lookup
```

---

## ERROR RECOVERY

```
MOUNT FAILED:
  → Verify /mnt/cpu/srv/aichat exists
  → Check network connection to CPU server
  → Retry: mount -c /mnt/cpu/srv/aichat /ai

SESSION NOT FOUND:
  → Create: ai "test"  (auto-creates session)
  → Verify: ls /ai/sessions/s$pid/

SLOW RESPONSES:
  → Reduce tokens: echo "tokens: 256" > ctl
  → Check: cat /ai/ctl (server status)
  → Try: ai "simple query" (test connectivity)

HISTORY NOT UPDATING:
  → Check: cat /ai/ctl (is server running?)
  → Verify: test -f /ai/sessions/s$pid/history
  → Debug: ls -la /ai/sessions/s$pid/
```

---

## FILE SIZES

```
Component           Typical Size    Notes
─────────────────────────────────────────────
aichat binary       2-5 MB          Includes libs
aichat.rc library   20 KB           All functions
Feature JSON        5-10 KB         Per feature
Session meta        <1 KB           Fixed size
History buffer      1-8 MB          256 entries
Context buffer      4-8 KB          Per session
```

---

## KEY NUMBERS

```
Metric                      Value       Config
─────────────────────────────────────────────
Max sessions                64          MAXSESSIONS
Max history entries         256         MAXHISTORY
Context buffer size         8 KB        MAXCONTEXT
Message buffer size         4 KB        MSGBUFSIZE
Token budget                512         token_limit
Default temperature         0.3         temperature
Process sleep on lock       -           QLock
Protocol version            9P2000      libpthread
```

---

## NETWORK TOPOLOGY

```
TERMINAL (User)
    │
    │ 9P Protocol (mount)
    │
    ▼
CPU SERVER
    ├─ aichat 9P Server
    │   ├─ Session Manager
    │   ├─ 9P Handler
    │   └─ Backend Processor
    │
    └─ External Services
        ├─ Claude API
        ├─ OpenAI API
        └─ Other LLM APIs

DISTRIBUTED FEATURES:
  ✓ Terminal in South Africa
  ✓ CPU Server in North America
  ✓ Transparent to user
  ✓ No special configuration
```

---

## ARCHITECTURE PRINCIPLES

```
1. EVERYTHING IS A FILE
   Sessions → directories
   State → readable files
   Queries → writeable files
   Results → file contents

2. PER-PROCESS ISOLATION
   Each process gets own session
   Kernel enforces boundaries
   No shared mutable state
   Automatic cleanup

3. DISTRIBUTION THROUGH 9P
   Works across networks
   Transparent to application
   Built-in authentication
   Capability-based security

4. COMPOSITION VIA UNION
   Features stackable
   Namespace flexible
   Standard tools work (ls, cat, grep)
   No special UI needed

5. SESSION AWARENESS
   Context accumulated
   History persistent
   Multi-turn natural
   State lifecycle = process lifetime
```

---

## QUICK START

```rc
# 1. Mount service
mount -c /mnt/cpu/srv/aichat /ai

# 2. Load functions
. /lib/aichat/aichat.rc

# 3. First query
ai 'explain Plan 9 mounting'

# 4. Interactive mode
aichat

# 5. Check session
cat /ai/sessions/s$pid/history
```

---

**Plan 9 + aichat = Distributed AI Assistant**
**Delivered January 18, 2026**
**Johannesburg, South Africa**