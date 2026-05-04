# RC Shell with AI Integration - User Guide

Welcome to the next evolution of Unix shell computing! This guide will help you master the RC shell enhanced with AI chat capabilities.

## 🚀 Quick Start

### Installation and Setup

1. **Build RC Shell:**
   ```bash
   make
   ```

2. **Install AI Chat (automatic):**
   ```bash
   bash install_aichat.sh
   ```

3. **Start Enhanced RC Shell:**
   ```bash
   ./rc_with_aichat.sh
   ```

4. **Verify Installation:**
   ```bash
   help                    # Shows integrated help system
   ai "hello world"        # Test AI integration
   ```

## 📚 RC Shell Fundamentals

### What Makes RC Different

RC is not just another shell—it's a revolutionary approach to command-line computing. Unlike bash or zsh, RC treats **everything as lists**, eliminating the complexity and bugs common in traditional shells.

### Core Concepts

#### 1. Lists Are Everything
```bash
# Simple list
files = (main.c utils.c parser.c)

# Lists from commands
headers = `{find . -name '*.h'}

# List operations
echo $#files           # Count: 3
echo $files(2)         # Second element: utils.c
echo $^files           # Flattened: "main.c utils.c parser.c"
```

#### 2. No Word Splitting Bugs
```bash
# Traditional shell problem:
# file="my document.txt"
# cp $file backup/     # BREAKS! Copies "my" and "document.txt"

# RC solution:
file = 'my document.txt'
cp $file backup/       # WORKS! No word splitting ever
```

#### 3. C-like Syntax
```bash
# Conditionals
if (~ $status 0) {
    echo 'Success!'
} else {
    echo 'Failed with status:' $status
}

# Loops
for (file in *.c) {
    echo 'Processing:' $file
    cc -c $file
}

# Functions
fn backup {
    for (file in $*) {
        cp $file $file.bak
        echo 'Backed up:' $file
    }
}
```

## 🤖 AI Integration Guide

### Basic AI Commands

#### General AI Assistant
```bash
ai "explain what this command does: tar -czf backup.tar.gz *.c"
ai "how do I find all files modified in the last 7 days?"
ai "write a shell function to monitor disk usage"
```

#### Shell-Specific Help
```bash
ask "how do I iterate over files in RC shell?"
ask "what's the difference between ~ and match in RC?"
ask "how do I handle errors in RC scripts?"
```

#### Learning Assistant
```bash
learn "RC shell list operations"
learn "pattern matching with ~ operator"
learn "RC shell functions vs bash functions"
```

#### System Administration
```bash
admin "monitor system resources"
admin "set up log rotation"
admin "check network connectivity"
```

#### Command Suggestions
```bash
suggest "backup all my C source files"
suggest "find duplicate files in directory"
suggest "monitor CPU usage continuously"
```

### AI Role System

The AI assistant has specialized roles for different contexts:

- **shell-expert**: Command-line scripting and automation
- **rc-tutor**: Learning RC shell specifically
- **sysadmin**: System administration tasks

```bash
# Use specific roles
ai -r shell-expert "optimize this pipeline: cat file | grep pattern | sort | uniq"
ai -r rc-tutor "explain RC's variable scoping rules"
ai -r sysadmin "set up automated backups"
```

## 🎯 Essential RC Shell Features

### Variables and Lists

#### Assignment and Access
```bash
# Simple assignment
name = 'John Doe'
numbers = (1 2 3 4 5)

# Multiple assignment
a = b = c = 'same value'

# Local assignment
PATH=. {
    # PATH is modified only for this block
    command_that_needs_current_dir
}
# PATH restored automatically
```

#### List Subscripting
```bash
args = (alpha beta gamma delta epsilon)

# Single subscript
echo $args(3)              # gamma

# Multiple subscripts
echo $args(1 3 5)          # alpha gamma epsilon

# Range subscripts (simulated)
echo $args(2-4)            # Use: $args(2 3 4)

# Count elements
echo $#args                # 5

# Check if empty
~ $#args 0 && echo 'empty' || echo 'not empty'
```

#### List Concatenation
```bash
# Pairwise concatenation
prefixes = (lib usr local)
suffixes = (bin lib include)
echo $prefixes^$suffixes   # libbin usrlib localinclude

# Distributive concatenation
cc -^(O g c) file1.c file2.c   # cc -O -g -c file1.c file2.c

# String building
host = server
port = 8080
url = http://$host:$port   # http://server:8080
```

### Pattern Matching

#### File Globbing
```bash
# Basic patterns
sources = *.c              # All C files
headers = *.[ch]           # C and header files
backups = *.bak           # Backup files

# Advanced patterns
tests = test_*.c          # Test files
configs = config.[0-9]*   # Numbered configs
```

#### String Matching with ~
```bash
# Test string patterns
~ $filename *.txt && echo 'Text file'
~ $user root admin && echo 'Privileged user'

# Test against multiple patterns
~ $file *.c *.h && echo 'C source code'

# List matching
files = (readme.txt config.ini data.csv)
~ $files *.txt && echo 'Contains text files'
```

#### Character Classes
```bash
# Standard ranges
~ $char [a-z] && echo 'lowercase letter'
~ $num [0-9] && echo 'digit'

# RC's tilde negation (cleaner than ^)
~ $char [~0-9] && echo 'not a digit'
~ $file [~.]* && echo 'not hidden file'
```

### Functions

#### Function Definition
```bash
# Simple function
fn greet {
    echo 'Hello,' $1'!'
}

# Function with multiple arguments
fn copy_with_backup {
    if (test -f $2) {
        cp $2 $2.bak
        echo 'Backed up existing file'
    }
    cp $1 $2
}

# Function with error handling
fn safe_remove {
    for (file in $*) {
        if (test -f $file) {
            rm $file && echo 'Removed:' $file
        } else {
            echo 'File not found:' $file
        }
    }
}
```

#### Function Examples
```bash
# Directory navigation with history
dirs = ()
fn pd {  # pushd
    dirs = ($pwd $dirs)
    cd $1
}

fn od {  # popd
    if (~ $#dirs 0) {
        echo 'Directory stack empty'
    } else {
        cd $dirs(1)
        dirs = $dirs(2-)  # Remove first element
    }
}

# Smart file search
fn ff {  # find file
    find . -name '*'^$1^'*' -type f
}

# Process management
fn killname {
    pids = `{ps aux | awk '$11 ~ /'$1'/ {print $2}'}
    if (~ $#pids 0) {
        echo 'No processes found matching:' $1
    } else {
        echo 'Killing processes:' $pids
        kill $pids
    }
}
```

### Control Structures

#### Conditional Execution
```bash
# Simple if
if (test -f config.txt) {
    echo 'Config file exists'
}

# If-else
if (~ $USER root) {
    echo 'Running as root'
} else {
    echo 'Running as user:' $USER
}

# Multiple conditions
if (test -f Makefile && test -x ./configure) {
    ./configure && make
} else {
    echo 'Not a buildable project'
}
```

#### Loops
```bash
# For loop with list
for (lang in C Python Go Rust) {
    echo 'I like' $lang
}

# For loop with command substitution
for (file in `{find . -name '*.txt'}) {
    echo 'Processing:' $file
    # Process file
}

# While loop
count = 1
while (test $count -le 10) {
    echo 'Count:' $count
    count = `{expr $count + 1}
}

# While with pattern
while (~ `{date +%S} *0) {
    echo 'Waiting for minute mark...'
    sleep 1
}
```

#### Switch Statements
```bash
switch ($file_ext) {
case c h
    echo 'C source code'
    cc -c $file
case py
    echo 'Python script'
    python $file
case go
    echo 'Go source'
    go run $file
case *
    echo 'Unknown file type'
}
```

### Advanced I/O

#### Redirection
```bash
# Basic redirection
command >output.txt 2>error.txt

# Merge streams
command >output.txt 2>&1

# Append mode
command >>logfile.txt

# Advanced descriptor manipulation
command >[2=1] >combined.log
```

#### Here Documents and Strings
```bash
# Here document
cat <<EOF
This is a multi-line
text block that will be
sent to cat
EOF

# Here string (immediate)
grep pattern <<<'search in this string'

# Variable substitution in here docs
name = 'World'
cat <<EOF
Hello, $name!
Your home is: $home
EOF
```

#### Process Substitution
```bash
# Compare sorted files
cmp <{sort file1.txt} <{sort file2.txt}

# Multi-way processing
data_stream | tee >{process_a} >{process_b} >{process_c}

# Complex pipelines
{
    echo 'Header'
    cat data.txt
    echo 'Footer'
} | tee logfile.txt | process_data
```

### Background Jobs and Process Management

#### Background Execution
```bash
# Run in background
long_running_task &
echo 'Started task with PID:' $apid

# Multiple background jobs
task1 &
task2 &
task3 &
echo 'All PIDs:' $apids

# Wait for specific job
wait $apid

# Wait for all jobs
wait
```

#### Signal Handling
```bash
# Custom signal handlers
fn sigint {
    echo 'Received interrupt - cleaning up...'
    cleanup_function
    exit 1
}

fn sigterm {
    echo 'Graceful shutdown requested'
    save_state
    exit 0
}

# Exit handler
fn sigexit {
    echo 'Shell exiting - goodbye!'
    cleanup_temp_files
}
```

## 🔧 Advanced Usage

### Complex Data Processing

#### Log Analysis
```bash
fn analyze_logs {
    for (log in /var/log/*.log) {
        echo 'Analyzing:' $log
        errors = `{grep ERROR $log | wc -l}
        warns = `{grep WARN $log | wc -l}
        echo 'Errors:' $errors 'Warnings:' $warns
    }
}
```

#### File System Operations
```bash
fn organize_downloads {
    for (file in ~/Downloads/*) {
        switch ($file) {
        case *.pdf
            mkdir -p ~/Documents/PDFs
            mv $file ~/Documents/PDFs/
        case *.jpg *.png *.gif
            mkdir -p ~/Pictures/
            mv $file ~/Pictures/
        case *.mp3 *.wav *.flac
            mkdir -p ~/Music/
            mv $file ~/Music/
        case *
            echo 'Unknown file type:' $file
        }
    }
}
```

#### System Monitoring
```bash
fn monitor_system {
    while (true) {
        clear
        echo '=== System Monitor ==='
        echo 'Time:' `{date}
        echo 'Load:' `{uptime | awk '{print $NF}'}
        echo 'Memory:' `{free -h | awk 'NR==2{print $3"/"$2}'}
        echo 'Disk:' `{df -h / | awk 'NR==2{print $5}'}
        sleep 5
    }
}
```

### AI-Assisted Development

#### Learn RC Progressively
```bash
# Start with basics
learn "RC shell variable assignment"
learn "RC shell list operations"
learn "RC shell function definition"

# Advanced topics
learn "RC shell pattern matching vs regex"
learn "RC shell signal handlers"
learn "RC shell memory management"
```

#### Get Contextual Help
```bash
# While writing scripts
ask "how do I check if a file exists in RC"
ask "RC shell equivalent of bash arrays"
ask "error handling patterns in RC shell"

# For system tasks
admin "set up cron job for backups"
admin "monitor disk space and alert"
admin "configure log rotation"
```

#### Command Optimization
```bash
# Get suggestions for improvement
suggest "make this faster: find . -name '*.c' | xargs grep 'TODO'"
suggest "parallelize this loop: for file in *; do process $file; done"
```

## 💡 Practical Examples

### Development Workflow

#### Project Build System
```bash
fn build {
    # AI-assisted build process
    suggestion = `{ai -r shell-expert "optimize build for C project with" $#sources "source files"}
    echo 'AI Suggestion:' $suggestion
    
    # Actual build
    if (test -f Makefile) {
        make clean && make -j`{nproc}
    } else if (test -f configure) {
        ./configure && make
    } else {
        echo 'No build system found'
        ask "how do I build a C project without Makefile?"
    }
}

fn test_suite {
    build || return
    
    # Run tests with AI monitoring
    results = `{make test 2>&1}
    if (~ $status 0) {
        echo 'All tests passed!'
    } else {
        echo 'Tests failed. Getting AI help...'
        ai -r shell-expert "analyze test failure:" $results
    }
}
```

#### Code Quality Tools
```bash
fn code_review {
    echo '=== Code Review Assistant ==='
    
    # Static analysis
    echo 'Running static analysis...'
    cppcheck_results = `{cppcheck --quiet . 2>&1}
    if (~ $#cppcheck_results 0) {
        echo 'No static analysis issues found'
    } else {
        echo 'Static analysis issues:'
        echo $cppcheck_results
        ai -r shell-expert "explain these cppcheck warnings:" $cppcheck_results
    }
    
    # Code metrics
    lines = `{find . -name '*.c' -exec wc -l {} + | tail -1}
    echo 'Total lines of code:' $lines(1)
    
    # Get AI suggestions
    ai -r shell-expert "suggest code quality improvements for C project with" $lines(1) "lines"
}
```

### System Administration

#### Automated Monitoring
```bash
fn smart_monitor {
    while (true) {
        # Collect metrics
        load = `{uptime | awk '{print $NF}' | tr -d ,}
        memory = `{free | awk 'NR==2{print int($3/$2*100)}'}
        disk = `{df / | awk 'NR==2{print int($5)}' | tr -d %}
        
        # Check thresholds with AI
        if (test $load -gt 2 || test $memory -gt 80 || test $disk -gt 90) {
            echo 'System alert detected!'
            admin "system showing high load:" $load "memory:" $memory"% disk:" $disk"%"
        }
        
        sleep 60
    }
}
```

#### Log Analysis
```bash
fn analyze_issues {
    echo 'Analyzing system logs...'
    
    # Extract recent errors
    errors = `{journalctl --since "1 hour ago" --priority err --no-pager}
    
    if (~ $#errors 0) {
        echo 'No recent errors found'
    } else {
        echo 'Recent errors detected:'
        echo $errors | head -10
        echo 'Getting AI analysis...'
        admin "analyze these system errors:" $errors
    }
}
```

#### Backup Management
```bash
fn intelligent_backup {
    backup_dir = /backup/`{date +%Y-%m-%d}
    
    # Ask AI for backup strategy
    strategy = `{admin "best backup strategy for home directory size:" `{du -sh $home}}
    echo 'AI Backup Strategy:' $strategy
    
    # Create backup
    mkdir -p $backup_dir
    rsync -av --progress $home/ $backup_dir/
    
    # Verify and report
    if (~ $status 0) {
        size = `{du -sh $backup_dir}
        echo 'Backup completed successfully:' $size
        admin "verify backup integrity for:" $backup_dir
    } else {
        echo 'Backup failed!'
        ask "how do I troubleshoot rsync backup failures?"
    }
}
```

### Daily Productivity

#### File Management
```bash
fn organize {
    # AI-assisted file organization
    for (file in *) {
        if (test -f $file) {
            category = `{ai "categorize this file: $file"}
            echo $file ': ' $category
            # Organize based on AI suggestion
        }
    }
}

fn smart_search {
    query = $1
    # Combine find with AI suggestions
    results = `{find . -iname '*'^$query^'*'}
    echo 'Found files:'
    echo $results
    
    if (~ $#results 0) {
        suggest "find files related to: $query"
    }
}
```

#### Task Automation
```bash
fn daily_tasks {
    echo '=== Daily Automation ==='
    
    # System cleanup
    echo 'Cleaning temporary files...'
    rm -rf /tmp/rc* >[2]/dev/null
    
    # Update package info
    echo 'Checking for updates...'
    if (which apt >/dev/null) {
        apt list --upgradable | head -5
    } else if (which brew >/dev/null) {
        brew outdated | head -5
    }
    
    # AI health check
    ai "analyze system health based on current time `{date} and uptime `{uptime}"
}
```

## 🛠 Configuration and Customization

### Environment Setup

#### Essential Variables
```bash
# In your .rcrc file
path = (/usr/local/bin /usr/bin /bin /opt/bin .)
cdpath = (. .. $home $home/projects)
history = $home/.rc_history

# AI configuration
OPENAI_API_KEY = 'your-api-key-here'
AICHAT_CONFIG_DIR = $home/.config/aichat
```

#### Custom Prompt
```bash
fn prompt {
    # Simple informative prompt
    current_dir = `{pwd | sed 's,^'$home',~,'}
    if (~ $status 0) {
        echo -n '['$current_dir']; '
    } else {
        echo -n '['$current_dir':'$status']; '
    }
}

# Or ask AI for creative prompts
fn setup_ai_prompt {
    ai_prompt = `{ai "design a creative shell prompt for RC shell"}
    echo 'AI suggested prompt:' $ai_prompt
}
```

### Advanced Configuration

#### Custom AI Roles
Edit `.aichat.yml` to add specialized roles:

```yaml
roles:
  - name: code-reviewer
    prompt: |
      You are a senior code reviewer. Analyze code for:
      - Security vulnerabilities
      - Performance issues
      - Best practices
      - Maintainability
      Provide specific, actionable feedback.

  - name: devops-expert
    prompt: |
      You are a DevOps engineer. Help with:
      - CI/CD pipeline setup
      - Infrastructure automation
      - Monitoring and alerting
      - Container orchestration
      Focus on practical, production-ready solutions.
```

#### Shell Function Library
```bash
# Create reusable function library
fn load_functions {
    if (test -f $home/.rc_functions) {
        . $home/.rc_functions
        echo 'Loaded custom functions'
    }
}

# Add to .rcrc
load_functions
```

## 🐛 Troubleshooting

### Common Issues

#### AI Integration Problems
```bash
# Test AI connectivity
fn test_ai {
    echo 'Testing AI integration...'
    result = `{ai "respond with: AI working" >[2=1]}
    if (~ $result *'AI working'*) {
        echo 'AI integration working ✓'
    } else {
        echo 'AI integration failed ✗'
        echo 'Result:' $result
        
        # Check common issues
        if (~ $OPENAI_API_KEY '') {
            echo 'Missing OPENAI_API_KEY environment variable'
        }
        
        if (! which aichat >/dev/null) {
            echo 'aichat binary not found in PATH'
        }
    }
}
```

#### RC Shell Issues
```bash
# Debug function definitions
fn debug_functions {
    echo 'Defined functions:'
    whatis -f
    
    echo 'Function issues:'
    for (func in `{whatis -f | awk '{print $2}'}) {
        if (! whatis $func >[2]/dev/null) {
            echo 'Broken function:' $func
        }
    }
}

# Debug variable scope
fn debug_vars {
    echo 'Current variables:'
    whatis -v | head -10
    
    echo 'Environment exports:'
    env | grep -E '^(PATH|HOME|USER)=' 
}
```

### Performance Optimization

#### Profile RC Performance
```bash
fn profile_rc {
    echo 'RC Performance Analysis'
    time {
        # Your RC operations here
        for (i in `{seq 1 1000}) {
            test_var = (item $i)
        }
    }
    
    # Get AI optimization suggestions
    ai -r shell-expert "optimize RC shell performance for list operations"
}
```

#### Memory Usage Monitoring
```bash
fn memory_profile {
    pid = $pid
    echo 'RC Shell Memory Usage:'
    ps -o pid,rss,vsz,comm -p $pid
    
    # Ask AI for memory optimization
    current_usage = `{ps -o rss= -p $pid}
    ai -r shell-expert "RC shell using" $current_usage "KB memory, suggest optimizations"
}
```

## 📖 Learning Path

### Beginner (First Week)
1. **Basic list operations**: Learn assignment, access, and counting
2. **Simple functions**: Create utility functions for daily tasks
3. **Pattern matching**: Master file globbing and string matching
4. **AI integration**: Use `ask` and `learn` commands regularly

### Intermediate (Second Week)
1. **Advanced I/O**: Process substitution and complex redirection
2. **Control structures**: Loops, conditionals, and error handling
3. **Variable scoping**: Local assignments and environment management
4. **AI assistance**: Use specialized roles and command suggestions

### Advanced (Third Week)
1. **Signal handling**: Custom signal handlers and process management
2. **Memory management**: Understand RC's arena allocation
3. **Performance optimization**: Profile and optimize RC scripts
4. **AI integration mastery**: Create custom workflows with AI

### Expert (Ongoing)
1. **Extend RC**: Add custom builtins and modifications
2. **Advanced patterns**: Complex glob patterns and matching
3. **System integration**: Deep OS integration and automation
4. **AI-driven development**: Use AI for code generation and review

## 🔗 Integration with Other Tools

### Version Control
```bash
fn git_ai_commit {
    changes = `{git diff --name-only}
    echo 'Changed files:' $changes
    
    # AI-generated commit message
    diff_summary = `{git diff --stat}
    commit_msg = `{ai -r shell-expert "write git commit message for changes:" $diff_summary}
    echo 'Suggested commit:' $commit_msg
    
    git add -A
    git commit -m $commit_msg
}
```

### Build Systems
```bash
fn intelligent_build {
    # Detect build system
    if (test -f Makefile) {
        build_cmd = make
    } else if (test -f build.gradle) {
        build_cmd = './gradlew build'
    } else if (test -f package.json) {
        build_cmd = 'npm run build'
    } else {
        ai_suggestion = `{ai "detect build system for this project"}
        echo 'AI suggestion:' $ai_suggestion
        return
    }
    
    echo 'Building with:' $build_cmd
    $build_cmd
}
```

### Documentation
```bash
fn generate_docs {
    # AI-assisted documentation
    for (source in *.c) {
        echo 'Documenting:' $source
        functions = `{grep '^[a-zA-Z].*(' $source | head -5}
        docs = `{ai -r code-reviewer "document these C functions:" $functions}
        echo $docs > $source.md
    }
}
```

## 🎓 Best Practices

### Script Organization
1. **Use functions**: Break complex scripts into functions
2. **Handle errors**: Check status and provide fallbacks
3. **Document with AI**: Use AI to generate documentation
4. **Test thoroughly**: Create test suites with AI assistance

### Performance Guidelines
1. **Avoid excessive subprocess creation**: Use RC builtins when possible
2. **Use lists efficiently**: Minimize concatenation operations
3. **Cache results**: Store expensive computations in variables
4. **Profile regularly**: Use time and AI analysis for optimization

### Security Considerations
1. **Validate inputs**: Always check user-provided data
2. **Use AI review**: Ask AI to review scripts for security issues
3. **Limit privileges**: Run with minimal necessary permissions
4. **Audit regularly**: Review and update security practices

## 📞 Getting Help

### Built-in Help System
```bash
help                    # Comprehensive help
help functions          # Specific topic help
help patterns           # Pattern matching guide
```

### AI-Powered Assistance
```bash
# Immediate help
ask "your specific question here"

# Learning mode
learn "topic you want to understand"

# Problem solving
admin "describe your system administration task"

# Command suggestions
suggest "describe what you want to accomplish"
```

### Community and Resources
- **GitHub Issues**: Report bugs and request features
- **AI Chat**: Get instant help and suggestions
- **Man Pages**: `man ./rc.1` for detailed reference
- **Examples**: Study `demo.rc` and `EXAMPLES` file

## 🎉 Welcome to the Future

You're now equipped with the most advanced shell environment available today. RC's elegant design combined with AI assistance creates an unparalleled command-line experience.

**Remember**: 
- Every operation is a list operation
- Pattern matching is your friend
- Functions are first-class citizens  
- AI is always ready to help
- The shell learns with you

**Happy computing!** 🚀

---

*For more advanced topics, use the `learn` command or consult the comprehensive documentation in the repository.*