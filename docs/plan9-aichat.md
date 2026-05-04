# Plan 9 AI Assistant Architecture for rc Shell
## A Distributed 9P-Native aichat Implementation

---

## Table of Contents

1. [System Architecture Overview](#system-architecture-overview)
2. [Feature Card Structure](#feature-card-structure)
3. [9P Distributed Setup](#9p-distributed-setup)
4. [Session Awareness Design](#session-awareness-design)
5. [rc Shell Integration](#rc-shell-integration)
6. [Implementation Examples](#implementation-examples)
7. [Security & Authentication](#security--authentication)

---

## System Architecture Overview

### Core Philosophy: Everything is a File

Plan 9's fundamental principle extends naturally to AI assistance. Rather than a traditional server-client RPC architecture, the Plan 9 aichat assistant exposes:

- **Assistant service** as a 9P file server mounted in `/ai`
- **Session state** as files within `/ai/sessions/{sessionid}`
- **Feature cards** as queryable file hierarchy
- **Command pipelines** as per-process namespaces

```
┌─────────────────────────────────────────────────────┐
│          Plan 9 Distributed Network                  │
├─────────────────────────────────────────────────────┤
│                                                       │
│  ┌──────────────┐    ┌──────────────┐               │
│  │  Terminal    │    │  CPU Server  │               │
│  │  /ai mounted │    │  Runs aichat │               │
│  └──────┬───────┘    └──────┬───────┘               │
│         │                    │                       │
│         │      9P Protocol   │                       │
│         └────────────────────┘                       │
│                    │                                  │
│         ┌──────────┴──────────┐                      │
│         │                     │                      │
│    ┌────┴──────┐      ┌──────┴────┐                 │
│    │ File Server│      │ Auth Server│                │
│    │/ai mounted │      │factotum    │                │
│    └────────────┘      └────────────┘                │
│                                                       │
└─────────────────────────────────────────────────────┘
```

---

## Feature Card Structure

### Card Anatomy (Plan 9 File Representation)

Each feature card is a logical grouping served as a queryable structure in `/ai/features/{featurename}/`.

#### JSON Card Model
```json
{
  "id": "shell-assist-001",
  "feature": "Shell Command Assistant",
  "category": "interactive",
  "description": "Suggests well-formed shell commands from natural language descriptions",
  "capabilities": [
    "Parse incomplete commands",
    "Fix malformed syntax",
    "Explain command intent",
    "Provide alternatives"
  ],
  "model_context": "You are a Plan 9 rc shell expert",
  "input_type": "natural_language",
  "output_type": "rc_command",
  "metadata": {
    "token_limit": 512,
    "temperature": 0.3,
    "created": "2026-01-18T07:11:00Z",
    "author": "aichat-p9"
  }
}
```

### File Hierarchy Representation

```
/ai/
├── features/
│   ├── shell-assist/
│   │   ├── meta           (card metadata)
│   │   ├── capabilities   (newline-separated list)
│   │   ├── prompt         (system prompt)
│   │   └── examples       (example exchanges)
│   │
│   ├── rc-explain/
│   │   ├── meta
│   │   ├── capabilities
│   │   ├── prompt
│   │   └── examples
│   │
│   └── code-review/
│       ├── meta
│       ├── capabilities
│       ├── prompt
│       └── examples
│
├── sessions/
│   ├── {sessionid}/
│   │   ├── meta           (session metadata)
│   │   ├── state          (session variables)
│   │   ├── history        (message log)
│   │   ├── context        (user context)
│   │   └── pid            (process binding)
│   │
│   └── {sessionid2}/
│       └── ...
│
├── ctl                     (control interface)
├── models                  (available models)
├── config                  (configuration)
└── stats                   (statistics)
```

### Feature Card Display (Text UI)

```
╔════════════════════════════════════════════════════════════╗
║ Shell Command Assistant                    [shell-assist]  ║
╠════════════════════════════════════════════════════════════╣
║                                                            ║
║ Parse natural language into rc shell commands with        ║
║ safety checks and Plan 9 idioms.                          ║
║                                                            ║
║ Capabilities:                                              ║
║  ✓ Command generation                                      ║
║  ✓ Syntax validation                                       ║
║  ✓ Plan 9 idiom compliance                                ║
║  ✓ Error explanation                                       ║
║                                                            ║
║ Model: claude-3.5-sonnet  | Context: 512 tokens           ║
║ Temperature: 0.3          | Safety: enabled                ║
║                                                            ║
╠════════════════════════════════════════════════════════════╣
║ [Use] [More Info] [Context] [History] [Tests]             ║
╚════════════════════════════════════════════════════════════╝
```

---

## 9P Distributed Setup

### aichat 9P Server Implementation (Pseudocode)

```c
#include <u.h>
#include <libc.h>
#include <auth.h>
#include <9p.h>
#include <thread.h>

/* Session structure */
typedef struct {
    char    id[32];           /* session ID */
    char    user[32];         /* authenticated user */
    ulong   pid;              /* process ID */
    ulong   ctime;            /* creation time */
    char    *history[256];    /* message history */
    int     histlen;
    char    *context;         /* session context */
    int     contextlen;
    Qid     qid;              /* Qid for this session */
} Session;

/* File server state */
typedef struct {
    Session *sessions[64];
    int     nsessions;
    QLock   sesslock;
    char    *models[16];      /* available models */
    int     nmodels;
} Aichat;

Aichat  aichat;
Srv     srv;

/* Qid type constants */
enum {
    Qroot = 0,
    Qctl,
    Qmodels,
    Qfeatures,
    Qsessions,
    Qsession,       /* specific session */
    Qsessmeta,
    Qsesshistory,
    Qsesscontext,
    Qsessctl,
};

/* Create session */
Session*
newsession(char *user, ulong pid)
{
    Session *s;
    char buf[64];
    
    s = emalloc(sizeof(Session));
    qlock(&aichat.sesslock);
    
    snprint(s->id, sizeof s->id, "s%lud", pid);
    strncpy(s->user, user, sizeof s->user - 1);
    s->pid = pid;
    s->ctime = time(nil);
    s->histlen = 0;
    s->context = emalloc(4096);
    s->contextlen = 0;
    
    aichat.sessions[aichat.nsessions++] = s;
    qunlock(&aichat.sesslock);
    
    return s;
}

/* Walk filesystem */
static void
aichawalk(Req *r)
{
    Fsinfo *f;
    int i;
    char *p;
    
    /* root directory */
    if(r->fid->qid.type == QTDIR && r->fid->qid.path == Qroot) {
        if(r->newfid->qid.path == Qctl)
            r->fid = r->newfid;
        else if(r->newfid->qid.path == Qmodels)
            r->fid = r->newfid;
        else if(r->newfid->qid.path == Qsessions)
            r->fid = r->newfid;
        else {
            respond(r, "path not found");
            return;
        }
        respond(r, nil);
    }
}

/* Read operation */
static void
aichread(Req *r)
{
    char *p;
    Session *s;
    int n;
    
    switch(r->fid->qid.path) {
    case Qroot:
        /* list root directory */
        r->ofcall.data = emalloc(512);
        r->ofcall.count = snprint(r->ofcall.data, 512,
            "ctl\nmodels\nfeatures\nsessions\n");
        respond(r, nil);
        break;
        
    case Qctl:
        /* control file - show server status */
        r->ofcall.data = emalloc(256);
        r->ofcall.count = snprint(r->ofcall.data, 256,
            "aichat server\nsessions: %d\nmodels: %d\n",
            aichat.nsessions, aichat.nmodels);
        respond(r, nil);
        break;
        
    case Qsessmeta:
        /* read session metadata */
        s = (Session*)r->fid->aux;
        if(s == nil) {
            respond(r, "invalid session");
            return;
        }
        r->ofcall.data = emalloc(256);
        r->ofcall.count = snprint(r->ofcall.data, 256,
            "id: %s\nuser: %s\npid: %lud\ntime: %lud\n",
            s->id, s->user, s->pid, s->ctime);
        respond(r, nil);
        break;
        
    case Qsesshistory:
        /* read session history */
        s = (Session*)r->fid->aux;
        if(s == nil) {
            respond(r, "invalid session");
            return;
        }
        /* return history lines */
        r->ofcall.data = emalloc(8192);
        r->ofcall.count = 0;
        for(int i = 0; i < s->histlen; i++) {
            n = snprint((char*)r->ofcall.data + r->ofcall.count,
                8192 - r->ofcall.count,
                "%s\n", s->history[i]);
            r->ofcall.count += n;
        }
        respond(r, nil);
        break;
    }
}

/* Write operation - submit queries */
static void
aichwrite(Req *r)
{
    Session *s;
    char query[1024];
    
    switch(r->fid->qid.path) {
    case Qsessctl:
        /* write command to session */
        s = (Session*)r->fid->aux;
        if(s == nil) {
            respond(r, "invalid session");
            return;
        }
        
        /* capture query */
        memmove(query, r->ifcall.data, 
            r->ifcall.count < sizeof query ? 
            r->ifcall.count : sizeof query - 1);
        query[r->ifcall.count] = 0;
        
        /* process with aichat backend */
        processquery(s, query);
        
        respond(r, nil);
        break;
    }
}

void
main(int argc, char *argv[])
{
    if(argc < 1) {
        fprint(2, "usage: aichat [mountpoint]\n");
        exits("usage");
    }
    
    /* Initialize aichat server */
    memset(&aichat, 0, sizeof aichat);
    aichat.nsessions = 0;
    
    /* Load available models */
    loadmodels();
    
    /* Set up 9P server */
    srv.walk = aichawalk;
    srv.read = aichread;
    srv.write = aichwrite;
    srv.auth = auth;
    srv.flush = nil;
    srv.attach = aichattach;
    srv.create = nil;
    srv.remove = nil;
    srv.stat = aichstat;
    
    /* Post service to /srv */
    postmountsrv(&srv, "aichat", argv[1], MBEFORE);
}
```

### Mounting the AI Service

```rc
# On terminal
9fs aide                          # mount CPU server
mount -c /mnt/aide/srv/aichat /ai

# Verify mount
ls /ai
# Output:
# ctl
# models
# features
# sessions

# Create new session
echo 'new' > /ai/ctl
# Kernel creates /ai/sessions/s12345 for you
```

---

## Session Awareness Design

### Session State Management

Sessions leverage Plan 9's per-process namespace to maintain context-aware interaction:

```rc
# rc script: aichat-session.rc
#!/usr/bin/env rc

# Get or create session
fn getsession {
    if(! test -e /ai/sessions/$pid/meta) {
        echo 'create '$pid > /ai/ctl
    }
    echo $pid
}

# Session context - persistent across shell lifetime
fn ai {
    sessid=`{getsession}
    sessdir=/ai/sessions/$sessid
    
    # Add to session context
    echo 'user: '$user >> $sessdir/context
    echo 'pwd: '$pwd >> $sessdir/context
    echo 'shell: '$0 >> $sessdir/context
    
    # Send query
    echo $* > $sessdir/ctl
    
    # Read response with history
    cat $sessdir/history | tail -1
}

# Save session state on logout
fn cleanup {
    if(! test -e /ai/sessions/$pid) {
        return
    }
    sessdir=/ai/sessions/$pid
    
    # Archive session
    cp $sessdir/history /tmp/aichat-history-$pid.log
    echo 'close' > $sessdir/ctl
}

# Register cleanup on exit
onint=(cleanup)

# Interactive REPL mode
fn airepl {
    sessdir=/ai/sessions/`{getsession}
    
    echo 'Entering AI REPL'
    echo 'Type "exit" to quit'
    
    while(true) {
        echo -n 'ai> '
        line=`{read}
        
        switch($line) {
        case 'exit'
            break
        case 'history'
            cat $sessdir/history
        case 'context'
            cat $sessdir/context
        case *
            echo $line > $sessdir/ctl
            cat $sessdir/history | tail -1
        }
    }
}

# Export functions
export ai airepl cleanup getsession
```

### Session State Files (per-process)

```
/ai/sessions/12345/
├── meta
│   id: s12345
│   user: alice
│   pid: 12345
│   created: 2026-01-18T07:15:30Z
│   last_activity: 2026-01-18T07:16:45Z
│   features: [shell-assist, code-review]
│
├── context
│   user: alice
│   pwd: /home/alice/projects/plan9
│   shell: /bin/rc
│   term: /dev/cons
│   prompt: cpu%
│   bindings: []
│
├── history
│   Q: convert bytes to MB
│   A: {echo 'scale=2; bytes / 1048576' | bc}
│   Q: how to read Plan 9 source?
│   A: cat /sys/src/cmd/rc/rc.c
│
├── state
│   feature: shell-assist
│   model: claude-3.5-sonnet
│   temperature: 0.3
│   max_tokens: 512
│
└── ctl (write-only)
   Accepts: feature selection, queries, state resets
```

### Session Lifecycle

```
┌─────────────────────────────────────────┐
│ User starts rc shell                    │
└────────────────┬────────────────────────┘
                 │
                 ▼
        ┌───────────────────┐
        │ Mount /ai service │
        └────────┬──────────┘
                 │
                 ▼
        ┌───────────────────────────────┐
        │ getsession creates /ai/       │
        │ sessions/{pid}/               │
        └────────┬──────────────────────┘
                 │
                 ▼
        ┌──────────────────────────┐
        │ Session state active     │
        │ (user can invoke ai)     │
        └────────┬─────────────────┘
                 │
                 ▼ (per invocation)
        ┌──────────────────────────┐
        │ Update context if needed │
        │ Write query to ctl       │
        │ Read response from hist  │
        └────────┬─────────────────┘
                 │
                 ▼ (on exit)
        ┌──────────────────────────┐
        │ cleanup() archives hist  │
        │ Closes session          │
        └──────────────────────────┘
```

---

## rc Shell Integration

### Feature Card Interactive Usage

```rc
# Feature 1: Shell Command Assistant
fn aicmd {
    # Invoke shell-assist feature
    ai-feature shell-assist 'show me how to extract tar.gz files'
}

# Feature 2: rc Script Explanation
fn aiexplain {
    # Explain current script
    cmd = $1
    ai-feature rc-explain 'explain this rc command: '$cmd
}

# Feature 3: Code Review
fn aireviewed {
    # Review rc script
    file = $1
    ai-feature code-review < $file
}

# Built-in helpers
fn ai-feature {
    feature=$1
    shift
    query=$*
    
    sessid=`{getsession}
    featdir=/ai/features/$feature
    
    # Validate feature exists
    if(! test -d $featdir) {
        echo 'Feature '$feature' not found' >&2
        return 1
    }
    
    # Send feature-specific query
    echo $query > /ai/sessions/$sessid/ctl
    
    # Display response
    tail -1 /ai/sessions/$sessid/history
}
```

### Named Pipes for Pipeline Integration

```rc
# Use 9P-based named pipes for multi-process AI
fn aipipe {
    sessid=`{getsession}
    pipename=/ai/sessions/$sessid/pipe-`{date -n}
    
    # Create bidirectional pipe
    echo 'mkpipe '$pipename > /ai/sessions/$sessid/ctl
    
    # Process can read/write through pipe
    # Writer side
    cat input.txt > $pipename &
    
    # Reader side  
    cat $pipename | grep pattern
}
```

### Union Mounting for Feature Access

```rc
# Bind all feature cards into process namespace
union /ai/features /features

# Now directly access features by path
ls /features                          # list all features
cat /features/shell-assist/meta       # read feature metadata
cat /features/shell-assist/prompt     # read system prompt
```

---

## Implementation Examples

### Example 1: Simple Shell Command Query

```rc
# Terminal session
cpu%

# Mount AI service (if not already mounted)
mount -c /mnt/cpu/srv/aichat /ai

# Get session ID
sessid=`{cat /ai/ctl | grep 'id:' | awk '{print $2}'}

# Query through session interface
echo 'find all *.go files modified in last 7 days' > /ai/sessions/$sessid/ctl

# Read response
cat /ai/sessions/$sessid/history
# Output:
# Q: find all *.go files modified in last 7 days
# A: find . -name '*.go' -mtime -7

# Execute suggested command
eval `{cat /ai/sessions/$sessid/history | tail -1 | sed 's/^A: //'}
```

### Example 2: Multi-Turn Conversation with Context

```rc
#!/usr/bin/env rc

# Multi-turn conversation maintaining context
fn aichat {
    sessdir=/ai/sessions/`{getsession}
    
    while(true) {
        echo -n 'You: '
        line=`{read}
        
        # Exit on empty or exit
        if(test -z $line || test $line = 'exit') {
            break
        }
        
        # Submit query with accumulated context
        {
            echo '# Previous context:'
            cat $sessdir/context
            echo '# New query:'
            echo $line
        } > $sessdir/ctl
        
        # Read AI response
        resp=`{cat $sessdir/history | tail -1}
        echo 'AI: '$resp
        
        # Update context for next turn
        echo 'context: '$line >> $sessdir/context
    }
}

aichat
```

### Example 3: Feature Card Display System

```c
/* Feature card renderer */
#include <u.h>
#include <libc.h>
#include <stdio.h>

typedef struct {
    char *title;
    char *desc;
    char **caps;
    int ncaps;
    char *model;
    char *temp;
} Card;

void
drawcard(Card *c)
{
    printf("╔════════════════════════════════════╗\n");
    printf("║ %s\n", c->title);
    printf("╠════════════════════════════════════╣\n");
    printf("║\n");
    printf("║ %s\n", c->desc);
    printf("║\n");
    printf("║ Capabilities:\n");
    for(int i = 0; i < c->ncaps; i++) {
        printf("║  ✓ %s\n", c->caps[i]);
    }
    printf("║\n");
    printf("║ Model: %s | Temp: %s\n", c->model, c->temp);
    printf("║\n");
    printf("╠════════════════════════════════════╣\n");
    printf("║ [Use] [Info] [History] [Settings]\n");
    printf("╚════════════════════════════════════╝\n");
}

void
main(void)
{
    Card shellAssist = {
        .title = "Shell Command Assistant",
        .desc = "Generate rc shell commands from natural language",
        .caps = (char*[]){"Syntax checking", "Plan 9 idioms", "Error messages"},
        .ncaps = 3,
        .model = "claude-3.5-sonnet",
        .temp = "0.3"
    };
    
    drawcard(&shellAssist);
}
```

---

## Security & Authentication

### Plan 9 Authentication with factotum

```rc
# Use Plan 9's authentication model
fn aiauthenticate {
    # Leverage factotum for protocol-agnostic authentication
    
    # Initialize authentication
    afd=`{fauth -n user /ai}
    
    if(test -z $afd) {
        echo 'Authentication failed' >&2
        return 1
    }
    
    # Use auth_proxy for secure connection
    auth_proxy $afd
}

# Mount with authentication
fn aimount {
    aiauthenticate
    mount -c /mnt/cpu/srv/aichat /ai
}
```

### Session Isolation via Per-Process Namespaces

```
Each rc shell process gets isolated session:

Process A (PID 1234)
  /ai/sessions/1234/  (only visible to Process A by default)
    ├── meta (user: alice)
    ├── history
    └── context

Process B (PID 5678)
  /ai/sessions/5678/  (only visible to Process B)
    ├── meta (user: bob)
    ├── history
    └── context

Isolation enforced by kernel at attach time
Authentication required to access other sessions
```

### Capability-Based Access Control

```
The 9P attach message includes credentials:
  - User name (from authentication)
  - Groups (from factotum)
  - Capabilities (permission tokens)

Server validates:
  - Can only read own session (/ai/sessions/{your-pid}/)
  - Can read public features (/ai/features/)
  - Cannot write other users' sessions
  - History is per-session and encrypted in transit
```

---

## Usage Workflow

### Complete Example: End-to-End Usage

```rc
#!/usr/bin/env rc

# 1. Start rc shell with AI support
export AICHAT_MOUNT=/ai

# 2. Mount service
mount -c /mnt/cpu/srv/aichat $AICHAT_MOUNT

# 3. Helper function
fn ai {
    sessdir=$AICHAT_MOUNT/sessions/`{getsession}
    echo $* > $sessdir/ctl
    cat $sessdir/history | tail -1
}

fn getsession {
    # Extract or create session ID from PID
    {echo 'session'; cat} | \
        awk 'NR==1 {print $NF}' 2>/dev/null || \
        echo $pid
}

# 4. Usage examples

# Simple one-shot query
ai 'list all .c files in current directory'

# Get explanation
ai 'explain what "union mounting" means in Plan 9'

# Code review
ai 'review this rc script for idioms' < script.rc

# Complex pipeline
{
    echo 'Generate a makefile for C project'
    ls *.c
} | ai

# Multi-turn session
airepl() {
    while(true) {
        echo -n 'ai> '
        line=`{read}
        test -z $line && break
        ai $line
    }
}

airepl
```

---

## Architecture Summary

| Component | Implementation | Location |
|-----------|-----------------|----------|
| **9P Server** | C with libc.h, 9p.h | /sys/src/cmd/aichat/ |
| **Session Manager** | Kernel /proc interface | /ai/sessions/{pid}/ |
| **Feature Cards** | JSON + text hierarchy | /ai/features/{name}/ |
| **Authentication** | factotum integration | /sys/auth/ |
| **rc Integration** | Shell functions | ~/.rcrc or /etc/profile |
| **Namespace** | Union mounting | Per-process /ai bind |

---

## Key Design Principles

1. **Everything is a File**: Sessions, features, and state are navigable through the file hierarchy
2. **Per-Process Namespaces**: Each shell gets isolated session context
3. **9P Native**: Direct protocol integration, no middleware layers
4. **Stateful Sessions**: Message history and context maintained across invocations
5. **Composable Features**: Feature cards can be combined via pipes and union mounts
6. **Authentication**: Leverages Plan 9's existing factotum infrastructure
7. **Distributed**: Works seamlessly across terminals, CPU servers, and file servers

---

## Files to Create

```
/sys/src/cmd/aichat/
├── aichat.c          (main 9P server)
├── session.c         (session management)
├── feature.c         (feature card handling)
├── auth.c            (authentication)
├── mkfile
├── aichat.1          (manual page)
└── examples/
    ├── aichat.rc     (rc shell integration)
    ├── airepl.rc     (interactive REPL)
    └── features/
        ├── shell-assist.json
        ├── code-review.json
        └── rc-explain.json
```

This architecture brings Plan 9's elegant distributed philosophy to modern AI assistants, maintaining the system's core principle that sophisticated functionality emerges from simple, composable file-based interfaces.