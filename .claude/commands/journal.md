Write a journal entry for today's work session on the `rem` project.

## Instructions

1. Determine today's date (YYYY-MM-DD format)
2. Check `journals/` dir for an existing file matching today's date pattern `*-<today's date>-journal.md`
3. **If a file for today exists**: Append a new entry AFTER the `---` separator at the bottom. Add a new `---` at the end of your entry.
4. **If no file for today exists**: Create a new file with the next sequence number: `<NNN>-<YYYY-MM-DD>-journal.md`. Add `---` at the end.

## Entry Format

```markdown
# Journal Entry <NNN> - <date> - <short title>
(or ## Session 2/3/etc if appending to existing file)

## Goal
<What was the objective this session?>

## What Changed
<Concrete list of what was built/fixed/refactored>

## Key Insights
<Technical learnings, gotchas discovered, things that failed and why, things that succeeded>

## Decisions Made
<Any architectural or design decisions and their rationale>

---
```

## Rules
- Be specific and technical - this is for future Claude sessions, not humans
- Include code snippets for non-obvious patterns
- Document what FAILED and why, not just successes
- Keep it concise but complete - no fluff
- Always end with `---` separator for future appending