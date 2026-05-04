
# RC Shell: A Deep Dive into the Plan 9 Shell Reimplementation

## 1. Introduction

The `rc` shell is a Unix reimplementation of the Plan 9 shell, originally designed by Tom Duff at Bell Labs. This version, primarily authored by Byron Rakitzis, brings the elegance and simplicity of the Plan 9 philosophy to the Unix environment. At its core, `rc` is a command interpreter that provides a clean, C-like syntax and a powerful, unified data model based on lists. This design eliminates many of the common pitfalls and complexities found in traditional Bourne-style shells, making it a more robust and predictable environment for both interactive use and scripting.

This document provides a comprehensive technical overview of the `rc` shell's architecture, core components, and internal workings. It is intended for developers, system architects, and anyone interested in the design and implementation of shell interpreters.

## 2. Core Architecture

The `rc` shell's architecture is designed for simplicity, robustness, and extensibility. It can be broken down into several key components that work together to parse, interpret, and execute commands.

### Key Components

The shell's functionality is modularized into several key components. The process begins with the **Lexical Analyzer** (`lex.c`), which transforms the raw input stream into a sequence of tokens. This component is responsible for identifying keywords, operators, and individual words, while also managing complex syntax like quoting and backslash escapes through a finite state automaton. Following lexical analysis, the **Parser** (`parse.y`) takes over, processing the token stream to construct an Abstract Syntax Tree (AST). This tree is a hierarchical representation of the command's structure, built according to the LALR(1) grammar defined in `parse.y`.

The core of the shell's execution logic resides in the **AST Walker/Executor** (`walk.c`). This component traverses the AST, interpreting the nodes and carrying out the specified actions, from simple command execution to the management of intricate control structures such as loops and conditionals. The shell’s state, including variables and functions, is managed by a dedicated set of components (`hash.c`, `var.c`, `fn.c`), which use hash tables for efficient storage and retrieval. All data within `rc` is fundamentally treated as a list of strings, and the **List and String Manipulation** module (`glom.c`) provides a rich set of functions for creating, expanding, and otherwise manipulating these lists.

To ensure efficient and robust operation, `rc` employs a custom **Memory Management** system (`nalloc.c`). This system is based on an arena allocator, which provides fast allocation and deallocation of memory for the lifetime of a single command, minimizing overhead and preventing memory leaks. The shell also includes a suite of **Built-in Commands** (`builtins.c`) for essential operations like `cd`, `exit`, `wait`, and `eval`. Finally, the architecture is rounded out by a sophisticated **Signal Handling** mechanism (`signal.c`) that allows for graceful management of interrupts and other system signals, and an **Input Handling** module (`input.c`) that manages a stack of input sources, enabling the shell to seamlessly read from files, strings, or the interactive terminal.


## 3. Data Structures

The `rc` shell relies on a set of well-defined data structures to represent its state and the commands it executes. The most important of these are the `Node` and `List` structures, which form the backbone of the shell's data model.

### The `List` Structure

In `rc`, the fundamental data type is the list of strings. The `List` structure is a simple singly-linked list, where each element contains a pointer to a string (`w`), a pointer to a metadata string (`m`) used for globbing, and a pointer to the next element in the list (`n`).

```c
struct List {
    char *w, *m;
    List *n;
};
```

This unified data model simplifies the shell's design by eliminating the distinction between scalars and arrays that is common in other shells. All variables, command arguments, and the results of command substitutions are represented as lists.

### The `Node` Structure

The `Node` structure is the building block of the Abstract Syntax Tree (AST). Each `Node` has a `type` field that indicates the kind of operation it represents (e.g., a command, a pipe, a conditional), and a union that holds pointers to other nodes or immediate values.

```c
struct Node {
    nodetype type;
    union {
        char *s;
        int i;
        Node *p;
    } u[4];
};
```

The `nodetype` enum defines all the possible types of nodes in the AST, such as `nPipe`, `nIf`, `nForin`, and `nAssign`.

### Other Key Data Structures

- **`Variable`**: Represents a shell variable, containing a pointer to its `List` value and its external string representation.
- **`rc_Function`**: Represents a shell function, containing a pointer to the AST of its body.
- **`Estack`**: A stack used for exception handling, which allows `rc` to manage control flow for `break`, `continue`, and `return` statements, as well as error conditions.


## 4. Command Execution

The execution of a command in `rc` is a multi-stage process that begins with parsing the input and ends with the command producing its output. This process is orchestrated by the interplay of the lexer, parser, and the AST walker.

### From String to AST

The first step in command execution is parsing. The **lexical analyzer** (`lex.c`) reads the raw command string and breaks it down into a series of tokens. These tokens represent the basic elements of the shell's grammar, such as words, operators, and keywords. The parser (`parse.y`), which is implemented using `yacc` (or `bison`), then takes this stream of tokens and constructs an **Abstract Syntax Tree (AST)**. The AST is a tree-like data structure that represents the grammatical structure of the command.

For example, a simple command like `ls -l /tmp` would be parsed into an AST that represents a command with a list of arguments.

![AST for a simple command](https://private-us-east-1.manuscdn.com/sessionFile/vqUygy1cz9iChG3YWjw6mx/sandbox/aNo5F0SGPvyWp2q5s3ip1B-images_1768739251017_na1fn_L2hvbWUvdWJ1bnR1L3JjX2FuYWx5c2lzL2FzdF9kaWFncmFt.png?Policy=eyJTdGF0ZW1lbnQiOlt7IlJlc291cmNlIjoiaHR0cHM6Ly9wcml2YXRlLXVzLWVhc3QtMS5tYW51c2Nkbi5jb20vc2Vzc2lvbkZpbGUvdnFVeWd5MWN6OWlDaEczWVdqdzZteC9zYW5kYm94L2FObzVGMFNHUHZ5V3AycTVzM2lwMUItaW1hZ2VzXzE3Njg3MzkyNTEwMTdfbmExZm5fTDJodmJXVXZkV0oxYm5SMUwzSmpYMkZ1WVd4NWMybHpMMkZ6ZEY5a2FXRm5jbUZ0LnBuZyIsIkNvbmRpdGlvbiI6eyJEYXRlTGVzc1RoYW4iOnsiQVdTOkVwb2NoVGltZSI6MTc5ODc2MTYwMH19fV19&Key-Pair-Id=K2HSFNDJXOU9YS&Signature=bSkyFxCu-lONIhpBBVCOsxFa6umcsorQYu~kjbFeYePNrAPUJWGsDmFQNjr-8~QKs-DaNDpIOmTbbd4s4I5RSm1M6hyjACVLBr1NHIdhYZY9wm78EU2vd7hRqaCNlNV3nDkSxY5P1cqIySMlB-5dAzFUSdGU7JXN7f8t0CqqN0oWqSE71CYfwX2OEkULdBhAgsiJLVvVL-nwonEZZccy2rSZzGr-sQp3MtyzZ4la635Th0JOD7g3dpTbidLmSoE07G-tScWF0QX4~qWYmGT8z7MjxdWkqnZVBLbrTTMZpYVupBZI13obHja25fGS6frSm-th85--43iuYnydoYuXaw__)

### Walking the AST

Once the AST is built, the `walk()` function in `walk.c` is called to execute it. This function recursively traverses the AST, executing the operations defined by each node. For a simple command node, `walk()` will call the `exec()` function, which is responsible for finding the command on the system, forking a new process, and executing the command.

For more complex structures like pipelines (`nPipe`), conditionals (`nIf`), or loops (`nForin`), the `walk()` function implements the corresponding logic, making recursive calls to `walk()` to execute the sub-commands.

The overall execution flow can be visualized as follows:

![Execution Flow Diagram](https://private-us-east-1.manuscdn.com/sessionFile/vqUygy1cz9iChG3YWjw6mx/sandbox/aNo5F0SGPvyWp2q5s3ip1B-images_1768739251018_na1fn_L2hvbWUvdWJ1bnR1L3JjX2FuYWx5c2lzL2V4ZWN1dGlvbl9mbG93.png?Policy=eyJTdGF0ZW1lbnQiOlt7IlJlc291cmNlIjoiaHR0cHM6Ly9wcml2YXRlLXVzLWVhc3QtMS5tYW51c2Nkbi5jb20vc2Vzc2lvbkZpbGUvdnFVeWd5MWN6OWlDaEczWVdqdzZteC9zYW5kYm94L2FObzVGMFNHUHZ5V3AycTVzM2lwMUItaW1hZ2VzXzE3Njg3MzkyNTEwMThfbmExZm5fTDJodmJXVXZkV0oxYm5SMUwzSmpYMkZ1WVd4NWMybHpMMlY0WldOMWRHbHZibDltYkc5My5wbmciLCJDb25kaXRpb24iOnsiRGF0ZUxlc3NUaGFuIjp7IkFXUzpFcG9jaFRpbWUiOjE3OTg3NjE2MDB9fX1dfQ__&Key-Pair-Id=K2HSFNDJXOU9YS&Signature=VbAPYnZFu~wLBvbYhv3BZ5f3z7SITFyCAYVJ4iMLuI49OE2gjNovsjpydK0SkrnOPdrno9HUqxhRFwNU8ONr7zRXTFITSE9MAy2f4Nd5Z9e-ak4uxVnCDXl2RulyTf2jORA39Ru6TKJl52haodq6zUolsvGxV9DVjXgTE6msWj8n3AwT22wVJN3ksGJtdMHD~rSjxBzh3N-WH5zjgRSh5oHkorooSAgU44S1fr3pUCHWHjIfDcw6w3LxC9GZJ6BGhOquIxcPzIVHjO1Qu14wlIzGrZ9nB04Xmx93hUiSp8BlFaq2dfee4t8P8Rq38Yl8RmGxQ0zil0JzoYwrlMbd1Q__)

### Pipelines and Redirection

`rc` handles pipelines and I/O redirection by manipulating file descriptors. When a pipeline is encountered, the shell creates a pipe and forks two child processes. The standard output of the first process is connected to the write end of the pipe, and the standard input of the second process is connected to the read end. This allows the output of one command to be seamlessly fed into the input of the next.

I/O redirection is handled by a queue of redirection nodes (`redirq`). Before a command is executed, the shell processes this queue, opening files and duplicating file descriptors as necessary to set up the desired input and output streams.

### Variable and Function Expansion

Before a command is executed, all variable references and command substitutions are expanded. The `glom()` function in `glom.c` is responsible for this process. It traverses the relevant parts of the AST, looks up variable values in the symbol table, executes back-quoted commands, and constructs the final list of arguments that will be passed to the command.

## 5. Memory Management

`rc` employs a sophisticated yet efficient memory management strategy centered around a custom arena allocator, implemented in `nalloc.c`. This approach is designed to optimize performance for the typical lifecycle of a shell command, where many small allocations are made and then freed all at once.

### Arena Allocation

The core of `rc`'s memory management is the `nalloc()` function, which allocates memory from a pre-allocated "arena" or block of memory. When a command begins execution, a new arena is created. All allocations for the command's data structures—such as AST nodes, lists, and strings—are made from this arena.

This has several advantages:

- **Speed**: Allocating memory from an arena is extremely fast, as it typically involves just incrementing a pointer. There is no need for the overhead of `malloc()` and `free()` for each individual allocation.
- **Locality of Reference**: Since all of a command's data is allocated contiguously in the same arena, this can improve cache performance.
- **Simplified Deallocation**: When the command is finished, the entire arena can be freed at once, or simply marked as available for the next command. This eliminates the need to track and free each individual allocation, which is a common source of memory leaks in C programs.

### The `nalloc` and `ealloc` Families

`rc` uses two families of allocation functions:

- **`nalloc()`**: This is the arena allocator, used for all memory that has the same lifetime as the current command. This includes the AST, lists, and temporary strings.
- **`ealloc()`**: This is a wrapper around the standard `malloc()` and `realloc()` functions. It is used for memory that needs to persist beyond the lifetime of a single command, such as the definitions of shell functions and the values of environment variables.

This dual-allocation strategy allows `rc` to balance the speed of arena allocation with the need for persistent data storage.

## 6. Conclusion

The `rc` shell is a testament to the power of simple, elegant design. By adhering to the Plan 9 philosophy of a unified data model and a clean, C-like syntax, it provides a powerful and robust environment for command-line interaction. Its internal architecture, from the arena-based memory management to the sophisticated list manipulation and exception handling, is a masterclass in building efficient and reliable systems software.

This document has provided a high-level overview of the `rc` shell's architecture and key components. For a deeper understanding, the reader is encouraged to explore the source code and the original Plan 9 documentation.
