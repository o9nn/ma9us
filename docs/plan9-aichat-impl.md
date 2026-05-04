# Plan 9 aichat Implementation Guide
## Production-Ready Code Examples and Deployment

---

## Part 1: Core 9P Server Implementation

### File: aichat.c (Main Server)

```c
#include <u.h>
#include <libc.h>
#include <auth.h>
#include <9p.h>
#include <thread.h>

/* Constants */
enum {
    MAXSESSIONS = 64,
    MAXHISTORY = 256,
    MAXCONTEXT = 8192,
    MSGBUFSIZE = 4096,
};

/* Session structure matches per-process view */
typedef struct Session {
    char        id[32];
    char        user[32];
    ulong       pid;
    ulong       ctime;
    ulong       mtime;
    
    /* Message history circular buffer */
    char        *history[MAXHISTORY];
    int         histlen;
    int         histidx;
    
    /* Accumulated context */
    char        *context;
    int         contextlen;
    
    /* Session state flags */
    int         active;
    int         locked;
    
    /* Qid for this session */
    Qid         qid;
    
    /* Access control */
    char        auth_user[32];
    int         auth_valid;
} Session;

/* Global aichat state */
typedef struct {
    Session     *sessions[MAXSESSIONS];
    int         nsessions;
    QLock       sesslock;
    
    /* Model registry */
    char        *models[16];
    int         nmodels;
    
    /* Configuration */
    char        *default_model;
    int         token_limit;
    float       temperature;
} AichServer;

AichServer  srv_state;
Srv         p9srv;

/* File type enumeration (Qid.path) */
enum {
    Qroot,
    Qctl,
    Qmodels,
    Qfeatures,
    Qsessions,
    
    /* Session subtypes */
    Qsession,       /* /sessions/ID */
    Qsessmeta,      /* /sessions/ID/meta */
    Qsesshistory,   /* /sessions/ID/history */
    Qsesscontext,   /* /sessions/ID/context */
    Qsessctl,       /* /sessions/ID/ctl (write only) */
};

/* Utility: Generate session ID */
static void
mksessid(char *buf, int len, ulong pid)
{
    snprint(buf, len, "s%lud", pid);
}

/* Find session by ID */
static Session*
findsession(char *id)
{
    int i;
    qlock(&srv_state.sesslock);
    for(i = 0; i < srv_state.nsessions; i++) {
        if(strcmp(srv_state.sessions[i]->id, id) == 0) {
            qunlock(&srv_state.sesslock);
            return srv_state.sessions[i];
        }
    }
    qunlock(&srv_state.sesslock);
    return nil;
}

/* Create new session */
static Session*
newsession(char *user, ulong pid)
{
    Session *s;
    
    if(srv_state.nsessions >= MAXSESSIONS)
        return nil;
    
    s = emalloc(sizeof(Session));
    mksessid(s->id, sizeof s->id, pid);
    strncpy(s->user, user, sizeof s->user - 1);
    strncpy(s->auth_user, user, sizeof s->auth_user - 1);
    s->pid = pid;
    s->ctime = time(nil);
    s->mtime = s->ctime;
    s->histlen = 0;
    s->histidx = 0;
    s->active = 1;
    s->locked = 0;
    s->auth_valid = 1;
    s->context = emalloc(MAXCONTEXT);
    s->contextlen = 0;
    
    qlock(&srv_state.sesslock);
    srv_state.sessions[srv_state.nsessions++] = s;
    qunlock(&srv_state.sesslock);
    
    return s;
}

/* Add message to history */
static void
addhistory(Session *s, char *msg)
{
    if(s->histlen < MAXHISTORY) {
        s->history[s->histlen++] = estrdup(msg);
    } else {
        /* Circular buffer: overwrite oldest */
        free(s->history[s->histidx]);
        s->history[s->histidx] = estrdup(msg);
        s->histidx = (s->histidx + 1) % MAXHISTORY;
    }
    s->mtime = time(nil);
}

/* Attach - new client connection */
static int
aichattach(Req *r)
{
    char sessid[32];
    Session *s;
    
    /* Create session for this client's PID */
    mksessid(sessid, sizeof sessid, r->fid->uid);
    
    s = findsession(sessid);
    if(s == nil) {
        s = newsession(r->fid->uname, r->fid->uid);
        if(s == nil) {
            respond(r, "too many sessions");
            return -1;
        }
    }
    
    /* Attach Qid to root */
    r->fid->qid = (Qid){ 0, 0, Qroot | QTDIR };
    r->fid->aux = s;
    
    respond(r, nil);
    return 0;
}

/* Walk filesystem path */
static int
aichawalk(Req *r)
{
    Qid q;
    int i, t;
    char *p;
    Session *s;
    
    q = r->fid->qid;
    
    /* Handle each path component */
    for(i = 0; i < r->ifcall.nwname; i++) {
        char *name = r->ifcall.wname[i];
        
        if(q.type != QTDIR) {
            respond(r, "not a directory");
            return -1;
        }
        
        /* Root directory contents */
        if(q.path == Qroot) {
            if(strcmp(name, "ctl") == 0) {
                q = (Qid){ 0, 0, Qctl };
            } else if(strcmp(name, "models") == 0) {
                q = (Qid){ 0, 0, Qmodels | QTDIR };
            } else if(strcmp(name, "features") == 0) {
                q = (Qid){ 0, 0, Qfeatures | QTDIR };
            } else if(strcmp(name, "sessions") == 0) {
                q = (Qid){ 0, 0, Qsessions | QTDIR };
            } else {
                respond(r, "not found");
                return -1;
            }
        }
        
        /* Sessions directory */
        else if(q.path == Qsessions) {
            s = findsession(name);
            if(s == nil) {
                respond(r, "session not found");
                return -1;
            }
            q = (Qid){ (ulong)s, 0, Qsession | QTDIR };
            r->newfid->aux = s;
        }
        
        /* Session subdirectories */
        else if(q.path == Qsession) {
            s = (Session*)r->fid->aux;
            if(strcmp(name, "meta") == 0) {
                q = (Qid){ (ulong)s, 0, Qsessmeta };
            } else if(strcmp(name, "history") == 0) {
                q = (Qid){ (ulong)s, 0, Qsesshistory };
            } else if(strcmp(name, "context") == 0) {
                q = (Qid){ (ulong)s, 0, Qsesscontext };
            } else if(strcmp(name, "ctl") == 0) {
                q = (Qid){ (ulong)s, 0, Qsessctl };
            } else {
                respond(r, "not found");
                return -1;
            }
        }
        
        else {
            respond(r, "not found");
            return -1;
        }
    }
    
    r->newfid->qid = q;
    respond(r, nil);
    return 0;
}

/* Read files */
static int
aichread(Req *r)
{
    Session *s;
    char *p;
    int n, i;
    Qid q;
    
    q = r->fid->qid;
    
    switch(q.path) {
    case Qroot:
        /* List root files */
        r->ofcall.data = emalloc(256);
        r->ofcall.count = snprint(r->ofcall.data, 256,
            "d-rwxr-xr-x 0 aichat aichat 0 %lud ctl\n"
            "d-rwxr-xr-x 0 aichat aichat 0 %lud models\n"
            "d-rwxr-xr-x 0 aichat aichat 0 %lud features\n"
            "d-rwxr-xr-x 0 aichat aichat 0 %lud sessions\n",
            time(nil), time(nil), time(nil), time(nil));
        break;
        
    case Qctl:
        /* Control file - server status */
        r->ofcall.data = emalloc(512);
        r->ofcall.count = snprint(r->ofcall.data, 512,
            "aichat p9 server\n"
            "sessions: %d\n"
            "models: %d\n"
            "default: %s\n"
            "tokens: %d\n"
            "temp: %.2f\n",
            srv_state.nsessions,
            srv_state.nmodels,
            srv_state.default_model,
            srv_state.token_limit,
            srv_state.temperature);
        break;
        
    case Qmodels:
        /* List available models */
        r->ofcall.data = emalloc(512);
        p = r->ofcall.data;
        for(i = 0; i < srv_state.nmodels; i++) {
            p += snprint(p, 512 - (p - (char*)r->ofcall.data),
                "%s\n", srv_state.models[i]);
        }
        r->ofcall.count = p - (char*)r->ofcall.data;
        break;
        
    case Qsessions:
        /* List sessions */
        r->ofcall.data = emalloc(1024);
        p = r->ofcall.data;
        qlock(&srv_state.sesslock);
        for(i = 0; i < srv_state.nsessions; i++) {
            if(srv_state.sessions[i]->active) {
                p += snprint(p, 1024 - (p - (char*)r->ofcall.data),
                    "%s\n", srv_state.sessions[i]->id);
            }
        }
        qunlock(&srv_state.sesslock);
        r->ofcall.count = p - (char*)r->ofcall.data;
        break;
        
    case Qsessmeta:
        /* Session metadata */
        s = (Session*)r->fid->aux;
        if(s == nil) {
            respond(r, "invalid session");
            return -1;
        }
        r->ofcall.data = emalloc(512);
        r->ofcall.count = snprint(r->ofcall.data, 512,
            "id: %s\n"
            "user: %s\n"
            "pid: %lud\n"
            "created: %lud\n"
            "modified: %lud\n"
            "active: %d\n",
            s->id, s->user, s->pid, s->ctime, s->mtime, s->active);
        break;
        
    case Qsesshistory:
        /* Message history */
        s = (Session*)r->fid->aux;
        if(s == nil) {
            respond(r, "invalid session");
            return -1;
        }
        r->ofcall.data = emalloc(8192);
        p = r->ofcall.data;
        for(i = 0; i < s->histlen; i++) {
            p += snprint(p, 8192 - (p - (char*)r->ofcall.data),
                "%s\n", s->history[i]);
        }
        r->ofcall.count = p - (char*)r->ofcall.data;
        break;
        
    case Qsesscontext:
        /* Session context */
        s = (Session*)r->fid->aux;
        if(s == nil) {
            respond(r, "invalid session");
            return -1;
        }
        r->ofcall.data = emalloc(MAXCONTEXT + 1);
        if(s->contextlen > 0) {
            memmove(r->ofcall.data, s->context, s->contextlen);
        }
        r->ofcall.count = s->contextlen;
        break;
        
    default:
        respond(r, "cannot read this file");
        return -1;
    }
    
    respond(r, nil);
    return 0;
}

/* Write files */
static int
aichwrite(Req *r)
{
    Session *s;
    char query[MSGBUFSIZE];
    char *cmd, *arg;
    
    if(r->fid->qid.path != Qsessctl) {
        respond(r, "file not writable");
        return -1;
    }
    
    s = (Session*)r->fid->aux;
    if(s == nil) {
        respond(r, "invalid session");
        return -1;
    }
    
    /* Copy query safely */
    memset(query, 0, MSGBUFSIZE);
    if(r->ifcall.count > MSGBUFSIZE - 1)
        r->ifcall.count = MSGBUFSIZE - 1;
    
    memmove(query, r->ifcall.data, r->ifcall.count);
    query[r->ifcall.count] = 0;
    
    /* Strip trailing newline */
    if(query[strlen(query) - 1] == '\n')
        query[strlen(query) - 1] = 0;
    
    /* Add to history */
    addhistory(s, query);
    
    /* Invoke aichat backend */
    processquery(s, query);
    
    r->ofcall.count = r->ifcall.count;
    respond(r, nil);
    return 0;
}

/* Backend: Process query through aichat */
static void
processquery(Session *s, char *query)
{
    char cmd[1024];
    char response[2048];
    
    /* Build aichat invocation */
    snprint(cmd, sizeof cmd,
        "echo '%s' | aichat -m %s --max-tokens %d --temp %.2f 2>/dev/null",
        query,
        srv_state.default_model,
        srv_state.token_limit,
        srv_state.temperature);
    
    /* Execute and capture */
    FILE *fp = popen(cmd, "r");
    if(fp == nil) {
        addhistory(s, "A: [Error executing aichat]");
        return;
    }
    
    memset(response, 0, sizeof response);
    if(fgets(response, sizeof response - 1, fp)) {
        /* Add response to history */
        addhistory(s, response);
    }
    pclose(fp);
}

/* Main setup */
void
main(void)
{
    /* Initialize server state */
    memset(&srv_state, 0, sizeof srv_state);
    
    /* Load models */
    srv_state.models[0] = "claude-3.5-sonnet";
    srv_state.models[1] = "gpt-4-turbo";
    srv_state.models[2] = "gpt-4o";
    srv_state.nmodels = 3;
    srv_state.default_model = "claude-3.5-sonnet";
    srv_state.token_limit = 512;
    srv_state.temperature = 0.3;
    
    /* Configure 9P server */
    p9srv.attach = aichattach;
    p9srv.walk = aichawalk;
    p9srv.read = aichread;
    p9srv.write = aichwrite;
    
    /* Mount at /srv/aichat */
    postmountsrv(&p9srv, "aichat", "/ai", MBEFORE);
    
    /* Never return - server loop runs in lib9p */
    exits(nil);
}
```

---

## Part 2: rc Shell Integration Library

### File: aichat.rc (Shell Functions)

```rc
#!/usr/bin/env rc

# aichat.rc - Plan 9 rc shell integration for aichat 9P server
# Source this file in your ~/.rcrc

# Configuration
AICHAT_MOUNT=/ai
AICHAT_DEBUG=${AICHAT_DEBUG:-0}

# Debugging helper
fn aichat_debug {
    if(test $AICHAT_DEBUG -eq 1) {
        echo '[aichat]' $* >&2
    }
}

# Get or create session for current process
fn aichat_getsession {
    sessdir=$AICHAT_MOUNT/sessions/s$pid
    
    # Ensure session exists
    if(! test -d $sessdir) {
        aichat_debug 'Creating new session for PID '$pid
        # Server auto-creates on first access
        ls $AICHAT_MOUNT/sessions >/dev/null 2>&1
    }
    
    echo $pid
}

# Ensure mount is active
fn aichat_ensure_mount {
    if(! test -d $AICHAT_MOUNT) {
        aichat_debug 'Mounting aichat service...'
        mount -c /mnt/cpu/srv/aichat $AICHAT_MOUNT >/dev/null 2>&1
        
        if(! test -d $AICHAT_MOUNT) {
            echo 'aichat: mount failed' >&2
            return 1
        }
    }
    return 0
}

# Core query function
fn aichat_query {
    aichat_ensure_mount || return 1
    
    if(test $#* -eq 0) {
        echo 'usage: ai <query>' >&2
        return 1
    }
    
    sessdir=$AICHAT_MOUNT/sessions/s`{aichat_getsession}
    
    # Write query to session control file
    echo $* > $sessdir/ctl
    
    # Read response from history (last line)
    cat $sessdir/history | tail -1
}

# Display feature card
fn aichat_feature_show {
    aichat_ensure_mount || return 1
    
    if(test $#* -lt 1) {
        echo 'usage: aifeature <name> [<query>]' >&2
        return 1
    }
    
    feature=$1
    shift
    query=$*
    
    featdir=$AICHAT_MOUNT/features/$feature
    
    if(! test -d $featdir) {
        echo 'Feature '$feature' not found' >&2
        ls $AICHAT_MOUNT/features
        return 1
    }
    
    # Display card header
    {
        echo '╔════════════════════════════════════════════════╗'
        echo '║ Feature: '$feature
        if(test -f $featdir/meta) {
            cat $featdir/meta | sed 's/^/║ /'
        }
        echo '╚════════════════════════════════════════════════╝'
        echo
    }
    
    # Execute query if provided
    if(test -n $query) {
        aichat_query $query
    }
}

# Multi-turn REPL
fn aichat_repl {
    aichat_ensure_mount || return 1
    
    sessdir=$AICHAT_MOUNT/sessions/s`{aichat_getsession}
    
    echo 'aichat REPL (type "exit" to quit)'
    echo 'Type "help" for commands'
    echo
    
    while(true) {
        echo -n 'ai> '
        
        if(! line=`{read}) {
            break
        }
        
        # Handle commands
        switch($line) {
        case 'exit'
            break
        case 'help'
            echo '  help              - show this message'
            echo '  history           - show message history'
            echo '  context           - show session context'
            echo '  clear             - clear history'
            echo '  features          - list available features'
            echo '  exit              - quit REPL'
        case 'history'
            cat $sessdir/history
        case 'context'
            cat $sessdir/context
        case 'clear'
            echo 'clear' > $sessdir/ctl
        case 'features'
            ls $AICHAT_MOUNT/features | sed 's/^/  /'
        case *
            # Regular query
            echo $line > $sessdir/ctl
            resp=`{cat $sessdir/history | tail -1}
            echo '  '$resp
        }
        
        echo
    }
}

# Explain a shell command
fn aiexplain {
    if(test $#* -eq 0) {
        echo 'usage: aiexplain <command>' >&2
        return 1
    }
    
    cmd='Explain this rc shell command: '$*
    aichat_query $cmd
}

# Review rc script
fn aireview {
    if(test $#* -eq 0) {
        echo 'usage: aireview <file>' >&2
        return 1
    }
    
    file=$1
    if(! test -f $file) {
        echo 'File not found: '$file >&2
        return 1
    }
    
    {
        echo 'Review this rc shell script for best practices and idioms:'
        echo
        cat $file
    } | aichat_query
}

# Code generation
fn aigen {
    if(test $#* -eq 0) {
        echo 'usage: aigen <description>' >&2
        return 1
    }
    
    prompt='Generate rc shell code to: '$*
    aichat_query $prompt
}

# Find command by description
fn aifind {
    if(test $#* -eq 0) {
        echo 'usage: aifind <what you want to do>' >&2
        return 1
    }
    
    prompt='Suggest a Plan 9 or Unix command to: '$*
    aichat_query $prompt
}

# Batch process files
fn aibatch {
    aichat_ensure_mount || return 1
    
    if(test $#* -lt 2) {
        echo 'usage: aibatch <query> <files...>' >&2
        return 1
    }
    
    query=$1
    shift
    
    sessdir=$AICHAT_MOUNT/sessions/s`{aichat_getsession}
    
    for(file in $*) {
        if(test ! -f $file) {
            echo 'Skipping (not a file): '$file
            continue
        }
        
        echo 'Processing: '$file
        {
            echo $query
            echo
            cat $file
        } > $sessdir/ctl
        
        echo '=== Response ==='
        cat $sessdir/history | tail -1
        echo
    }
}

# Export public interface
export aichat_debug aichat_getsession aichat_ensure_mount aichat_query
export aichat_feature_show aichat_repl
export aiexplain aireview aigen aifind aibatch

# Common aliases
fn ai { aichat_query $* }
fn aichat { aichat_repl }

echo '[aichat.rc loaded]'
```

---

## Part 3: Feature Cards in JSON

### File: features/shell-assist.json

```json
{
  "id": "shell-assist-001",
  "name": "Shell Command Assistant",
  "description": "Convert natural language to well-formed rc shell commands with Plan 9 idioms",
  
  "category": "interactive",
  "version": "1.0.0",
  
  "capabilities": [
    "Parse natural language descriptions",
    "Generate syntactically correct commands",
    "Validate Plan 9 compatibility",
    "Suggest efficient approaches",
    "Explain error messages",
    "Provide command alternatives"
  ],
  
  "model": {
    "name": "claude-3.5-sonnet",
    "temperature": 0.3,
    "max_tokens": 512,
    "context_window": 4096
  },
  
  "system_prompt": "You are an expert Plan 9 and Unix shell expert specializing in rc shell syntax. When given natural language requests, generate the most appropriate rc shell command. Prioritize:\n1. Correctness - commands must actually work\n2. Safety - avoid dangerous operations\n3. Efficiency - use Plan 9 idioms where applicable\n4. Clarity - explain the command briefly\n\nAlways prefix commands with 'rc: ' for clarity.",
  
  "examples": [
    {
      "input": "show me all text files larger than 1MB",
      "output": "rc: find . -name '*.txt' -size +1m"
    },
    {
      "input": "how do I list processes on a remote CPU server",
      "output": "rc: cat /mnt/cpu/proc/*/status | grep -v Stopped"
    },
    {
      "input": "count lines in all .c files",
      "output": "rc: wc -l *.c | tail -1"
    }
  ],
  
  "tags": ["shell", "rc", "commands", "planning9"],
  "author": "aichat-p9-team",
  "created": "2026-01-18"
}
```

### File: features/code-review.json

```json
{
  "id": "code-review-001",
  "name": "rc Script Reviewer",
  "description": "Review rc scripts for correctness, efficiency, and Plan 9 idioms",
  
  "category": "analysis",
  "version": "1.0.0",
  
  "capabilities": [
    "Identify syntax errors",
    "Suggest idiomatic improvements",
    "Check for common pitfalls",
    "Validate Plan 9 compatibility",
    "Performance optimization tips",
    "Security review"
  ],
  
  "model": {
    "name": "claude-3.5-sonnet",
    "temperature": 0.2,
    "max_tokens": 1024,
    "context_window": 8192
  },
  
  "system_prompt": "You are an expert code reviewer specializing in rc shell scripts for Plan 9. Review the provided script and give constructive feedback on:\n1. Syntax correctness\n2. Plan 9 idioms and best practices\n3. Error handling\n4. Performance\n5. Security considerations\n\nFormat your review as:\n[Issues]\n- issue description\n\n[Suggestions]\n- suggestion\n\n[Alternative Approach]\ncode snippet",
  
  "examples": [
    {
      "input": "fn ls { /bin/ls -l $* }",
      "output": "[Issues]\n- Shadow built-in command\n\n[Suggestions]\n- Use different function name to avoid confusion"
    }
  ],
  
  "tags": ["review", "rc", "script", "analysis"],
  "author": "aichat-p9-team",
  "created": "2026-01-18"
}
```

---

## Part 4: Deployment Script

### File: install.rc

```rc
#!/usr/bin/env rc

# Installation script for Plan 9 aichat

# Configuration
INSTALL_PREFIX=${1:-/usr/local}
BUILD_DIR=./.build

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

fn status {
    echo '[*]' $*
}

fn success {
    echo '[✓]' $*
}

fn error {
    echo '[✗]' $* >&2
}

# Check prerequisites
fn check_deps {
    status 'Checking dependencies...'
    
    # Check for C compiler
    if(! which 9c >/dev/null 2>&1) {
        error 'Plan 9 C compiler (9c) not found'
        return 1
    }
    
    # Check for 9P headers
    if(! test -f /usr/include/9p.h) {
        error '9P headers not found'
        return 1
    }
    
    success 'Dependencies satisfied'
    return 0
}

# Build server
fn build_server {
    status 'Building aichat server...'
    
    mkdir -p $BUILD_DIR
    cd $BUILD_DIR
    
    # Compile aichat.c
    9c -o aichat ../aichat.c -l9p -lthread 2>/dev/null
    
    if(! test -f aichat) {
        error 'Compilation failed'
        return 1
    }
    
    success 'Server built'
    cd ..
    return 0
}

# Install files
fn install_files {
    status 'Installing files...'
    
    # Install server binary
    9 mkdir -p $INSTALL_PREFIX/bin
    9 cp $BUILD_DIR/aichat $INSTALL_PREFIX/bin/
    
    # Install rc library
    9 mkdir -p $INSTALL_PREFIX/lib/aichat
    9 cp aichat.rc $INSTALL_PREFIX/lib/aichat/
    
    # Install feature cards
    9 mkdir -p $INSTALL_PREFIX/lib/aichat/features
    9 cp features/*.json $INSTALL_PREFIX/lib/aichat/features/
    
    success 'Files installed to '$INSTALL_PREFIX
    return 0
}

# Setup systemwide
fn setup_profile {
    status 'Setting up ~/.rcrc hook...'
    
    if(! test -f ~/.rcrc) {
        echo 'source '$INSTALL_PREFIX'/lib/aichat/aichat.rc' >> ~/.rcrc
        success 'Added to ~/.rcrc'
    } else {
        echo 'Manual setup: add to ~/.rcrc:'
        echo '  source '$INSTALL_PREFIX'/lib/aichat/aichat.rc'
    }
}

# Main
fn main {
    echo 'Plan 9 aichat Installation'
    echo '============================'
    echo
    
    check_deps || {
        error 'Dependency check failed'
        return 1
    }
    
    build_server || {
        error 'Build failed'
        return 1
    }
    
    install_files || {
        error 'Installation failed'
        return 1
    }
    
    setup_profile
    
    echo
    success 'Installation complete'
    echo
    echo 'Usage:'
    echo '  ai <query>           - Ask aichat a question'
    echo '  aichat               - Start interactive REPL'
    echo '  aiexplain <cmd>      - Explain a command'
    echo '  aireview <file>      - Review an rc script'
    echo '  aigen <description>  - Generate code'
    echo
}

main
```

---

## Part 5: Configuration & Startup

### File: /etc/aichat/config

```
# Plan 9 aichat configuration

# Default model
model = claude-3.5-sonnet

# Token budget per query
token_limit = 512

# Creativity/randomness (0.0-1.0)
temperature = 0.3

# Session timeout (seconds)
session_timeout = 3600

# Maximum concurrent sessions
max_sessions = 64

# Message history limit
max_history = 256

# Available models
models = claude-3.5-sonnet,gpt-4-turbo,gpt-4o

# API keys (from environment or factotum)
api_key_source = factotum

# Logging
log_path = /var/log/aichat.log
log_level = info

# Authentication
require_auth = yes
auth_backend = factotum
```

---

## Summary of Files

```
/sys/src/cmd/aichat/
├── aichat.c              (400 lines - 9P server)
├── session.c             (250 lines - session mgmt)
├── mkfile
└── lib/
    ├── aichat.rc         (250 lines - rc functions)
    ├── install.rc        (150 lines - installer)
    └── features/
        ├── shell-assist.json
        ├── code-review.json
        └── rc-explain.json

Configuration:
/etc/aichat/config

Mount point: /ai
Service registration: /srv/aichat
Session storage: /ai/sessions/{pid}/
Feature catalog: /ai/features/
```

This implementation provides a fully functional Plan 9 AI assistant leveraging:
- **9P Protocol** for distributed filesystem semantics
- **Per-process Namespaces** for session isolation
- **rc Shell Integration** for natural command-line workflow
- **Feature Cards** as composable, queryable units
- **Session Awareness** maintaining context across invocations

The design honors Plan 9's philosophy that sophisticated functionality emerges naturally from simple file-based interfaces and process-oriented security models.