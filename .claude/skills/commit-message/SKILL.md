---
name: commit-message
description: Write a git commit message from the staged changes. Use when asked to generate, write, or draft a commit message. Produces a plain message (NOT Conventional Commits) with no AI/Claude/model attribution.
argument-hint: [optional extra context]
allowed-tools: Bash(git diff:*), Bash(git status:*), Bash(git log:*)
---

# Generate a git commit message

Write a commit message describing the staged changes. Output **only** the message text — no preamble, no explanation, no code fences.

## Steps

1. Inspect what is staged:
   - `git diff --cached` for the actual changes.
   - `git status --short` to see the set of files.
   - If nothing is staged, say so and offer to describe the unstaged changes instead (`git diff`).
2. Optionally skim recent history for phrasing/style: `git log --oneline -10`.
3. Write the message following the format and rules below. Fold in any extra context passed as an argument.

## Format

- **Summary line**: imperative mood ("Add", "Fix", "Remove", not "Added"/"Adds"), capitalized, no trailing period, aim for ≤ 50 characters (hard limit 72).
- **Body** (only when the change needs it): one blank line after the summary, then wrap at ~72 characters. Explain *what* changed and *why* — not a restatement of the diff. Use `-` bullets for multiple distinct points.
- Small, self-explanatory changes may be a summary line alone.

## Rules

- **Do NOT use Conventional Commits.** No `type:` / `type(scope):` prefixes — no `feat:`, `fix:`, `chore:`, `refactor:`, `docs:`, etc. Just write a plain descriptive summary.
- **Do NOT mention Claude, Anthropic, AI, assistants, or any model names** anywhere in the message.
- **Do NOT add attribution trailers** — no `Co-Authored-By:`, no `Generated with`, no session links, no `🤖` markers.
- Describe the change itself, not the process of making it.
- Match the tone and specificity of the repository's existing history.

## Output

Print the finished commit message and nothing else, ready to paste into `git commit`.
