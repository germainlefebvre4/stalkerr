## 1. Add the Architecture section

- [ ] 1.1 In `README.md`, insert a new `## Architecture` heading immediately after the `## Features` section and before `## Quick Start`, with a one-sentence intro line, and verify with `grep -n "^## "` that `Architecture` now appears between `Features` and `Quick Start` in heading order.
- [ ] 1.2 Add the "System context" Mermaid diagram (design.md, Diagram 1) under a `### System context` subheading, copied verbatim from `design.md`, and verify the fenced block opens with ` ```mermaid ` and closes with ` ``` ` (balanced fences).
- [ ] 1.3 Add the "Deployment topology" Mermaid diagram (design.md, Diagram 2) under a `### Deployment topology` subheading, copied verbatim from `design.md`, with the same fence-balance check.
- [ ] 1.4 Add the "`process` pipeline" Mermaid diagram (design.md, Diagram 3) under a `### The "process" pipeline` subheading, copied verbatim from `design.md`, with the same fence-balance check.
- [ ] 1.5 Add the "`download` pipeline" Mermaid diagram (design.md, Diagram 4) under a `### The "download" pipeline` subheading, copied verbatim from `design.md`, with the same fence-balance check.
- [ ] 1.6 Add a closing line linking to `docs/DATABASE.md#entity-relationships` for the data-model diagram, and verify the relative link path resolves (file and anchor both exist).

## 2. Verify

- [ ] 2.1 Render `README.md` in a Mermaid-aware previewer (GitHub PR preview, VS Code Markdown preview, or the Mermaid Live Editor pasted per-diagram) and confirm all four diagrams render without syntax errors.
- [ ] 2.2 Re-read the full new section once rendered and confirm no existing README content (Features, Quick Start, CLI Commands, API Documentation, Configuration sections) was altered or reordered - diff should show only additions.
