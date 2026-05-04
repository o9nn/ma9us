# 9fs9rc

**9P Protocol, rc Shell, and AI Chat Integration for Plan 9 Distributed Computing**

This repository contains the foundational components for integrating Plan 9's distributed computing model with modern AI capabilities. It brings together the elegance of the 9P protocol, the power of the rc shell, and a novel architecture for AI chat integration—all designed with AGI-OS principles in mind.

## Repository Structure

```
9fs9rc/
├── 9p/                     # 9P Protocol implementations
│   ├── u9fs/               # Unix 9P file server (9P2000.u)
│   └── 9p-rfc-master/      # 9P RFC specifications (XML)
├── rc/                     # rc shell implementation
│   └── rc/                 # Byron Rakitzis's Unix rc shell
├── docs/                   # Documentation
│   ├── 9p_documentation.md # 9P protocol technical overview
│   ├── rc_documentation.md # rc shell technical overview
│   ├── plan9-aichat.md     # AI chat integration design
│   ├── plan9-aichat-impl.md# Implementation details
│   ├── plan9-aichat-guide.md# User guide
│   ├── quick-ref-card.md   # Quick reference
│   └── hensbergen.pdf      # "Grave Robbers from Outer Space" paper
├── specs/                  # Z++ Formal Specifications
│   ├── 9p.z                # 9P2000 protocol formal spec
│   └── rc.z                # rc shell formal spec
└── diagrams/               # Architecture diagrams
    ├── 9p_flow.png         # 9P message sequence diagram
    ├── aichat_architecture.png # AI chat system architecture
    ├── ast_diagram.png     # rc shell AST structure
    └── execution_flow.png  # rc shell execution flow
```

## Components

### 9P Protocol (`9p/`)

The **9P2000** protocol is the cornerstone of Plan 9's "everything is a file" philosophy. It provides a simple, efficient, and powerful way to access resources across a network.

- **u9fs**: A user-space 9P file server for Unix systems, implementing the `9P2000.u` extension for Unix compatibility (symbolic links, device files, etc.)
- **9P RFC**: The formal protocol specification in XML format

### rc Shell (`rc/`)

The **rc shell** is Plan 9's command interpreter, reimplemented for Unix by Byron Rakitzis. It features:

- Clean, C-like syntax
- Unified list-based data model
- First-class functions and closures
- Powerful I/O redirection and pipelines

### AI Chat Integration (`docs/plan9-aichat*.md`)

A novel architecture for exposing AI assistants as 9P file systems, enabling:

- Shell-native AI interaction via `rc` functions
- Session management through synthetic files
- Integration with multiple LLM backends (Claude, GPT-4, etc.)

## Formal Specifications

The `specs/` directory contains **Z++ formal specifications** that precisely define:

- **9p.z**: 9P2000 message types, state schemas, and protocol operations
- **rc.z**: rc shell data structures, AST nodes, and execution semantics

## Key Papers

- **"Grave Robbers from Outer Space: Using 9P2000 Under Linux"** (Hensbergen & Minnich, USENIX 2005) - The foundational paper describing the 9P2000.u extension and u9fs implementation.

## Architecture Diagrams

| Diagram | Description |
|---------|-------------|
| `9p_flow.png` | 9P message sequence from version negotiation to file I/O |
| `aichat_architecture.png` | AI chat system component architecture |
| `ast_diagram.png` | rc shell Abstract Syntax Tree structure |
| `execution_flow.png` | rc shell command execution pipeline |

## Integration with AGI-OS

This repository is part of the broader **9cog** ecosystem, which aims to integrate:

- **Plan 9/Inferno** distributed operating system concepts
- **OpenCog** cognitive architecture
- **Tensor computing** frameworks (ATen/PyTorch)

The 9P protocol serves as the communication backbone, enabling cognitive processes to be exposed as file systems and accessed uniformly across distributed nodes.

## Building

### u9fs

```bash
cd 9p/u9fs
make
```

### rc shell

```bash
cd rc/rc
./configure
make
```

## References

1. Pike, R. et al. "Plan 9 from Bell Labs" - USENIX Computing Systems, 1995
2. Hensbergen, E. & Minnich, R. "Grave Robbers from Outer Space" - USENIX FREENIX, 2005
3. Duff, T. "Rc — The Plan 9 Shell" - Plan 9 Programmer's Manual

## License

Components retain their original licenses:
- u9fs: MIT/X11
- rc: Public Domain
- Documentation: CC BY-SA 4.0
