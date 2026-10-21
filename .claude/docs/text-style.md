# Text style enforcement

Rules for every string the agent writes - source code, log messages, comments, UI copy, doc files, commit messages.

## Forbidden punctuation

### Em-dash `—` (U+2014)

**Never use.** Anywhere. Period.

- **Why**: in the past, em-dashes have surfaced as mojibake (`â`, `â`) in log rows and log dashboards whenever a downstream consumer assumed Latin-1 or stripped non-ASCII. Hyphen-minus `-` (U+002D) has no such failure mode and reads the same in monospaced log viewers.
- **Where the rule applies**:
  - Go / TS / Svelte / proto source (strings AND comments)
  - `slog` / panic / error messages
  - UI copy (labels, buttons, headings, toast text, validation messages)
  - Markdown docs under `docs/`, `.claude/docs/`, top-level `*.md`
  - Commit messages and PR descriptions
- **How to apply**:
  - Where you would have written ` - ` (em-dash for parenthetical aside), use ` - ` (space + hyphen + space).
  - Where the em-dash separated cause from effect (`X - Y`), prefer rephrasing into two sentences when natural; otherwise the same ` - ` works.
  - Never combine two hyphens `--` to simulate an em-dash; that's a different keyboard artifact with its own pitfalls in CLI argument parsing.

### En-dash `–` (U+2013)

**Same rule: never use.** Same mojibake risk class. Use plain hyphen `-` for numeric ranges (`2020-2025`, `port 5000-5010`) and word joins.

## Why this is enforced as a separate rule

A blanket "use ASCII only" rule is too strict (the UI and docs are written in Korean). But em/en-dashes are *typographic* embellishments, never load-bearing - they survive the transition to a plain hyphen without ambiguity. So the line we draw is: non-ASCII is allowed when it carries information (proper names, identifiers, foreign-language UI strings), but typographic substitutes for plain ASCII are not.

## Exception

This file (`.claude/docs/text-style.md`) is the **only** place in the repo where U+2014 and U+2013 are allowed - they must appear here in the section headings above and inside the `sed` command below so the rule is self-documenting. Exclude `.claude/docs/text-style.md` from any grep-based audit.

## Detection

To audit the repo for regressions:

```sh
grep -rn $'—\|–' \
  --include="*.go" --include="*.ts" --include="*.tsx" --include="*.svelte" \
  --include="*.js" --include="*.md" --include="*.proto" --include="*.yml" --include="*.yaml" \
  . | grep -v "proto-gen\|node_modules\|\.svelte-kit\|\.git\|vendor"
```

Bulk-fix:

```sh
# files to clean (extend the find pattern as needed)
find . -type f \( -name "*.go" -o -name "*.ts" -o -name "*.tsx" -o -name "*.svelte" \
                 -o -name "*.js" -o -name "*.md" -o -name "*.proto" -o -name "*.yml" -o -name "*.yaml" \) \
  -not -path "*/node_modules/*" -not -path "*/.svelte-kit/*" -not -path "*/proto-gen/*" \
  -not -path "*/build/*" -not -path "*/.git/*" -not -path "*/vendor/*" \
  | xargs sed -i $'s/—/-/g; s/–/-/g'
```
