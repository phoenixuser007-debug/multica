# Fix-generation framing — ported from the original Python pipeline

The original pipeline asked gpt-4o for a JSON array of change blocks instead of a raw unified diff — this sidestepped the LLM's tendency to emit unified-diff syntax that `git apply` rejected. You (Claude) should follow the same pattern: reason about the minimal change, then **apply it directly** with the Edit tool (preferred) or `sed`, and let `git diff` produce the authoritative patch.

## Output shape (if you want to stage changes as JSON first)

```json
[
  {
    "old_code": "if (reporter != null) {\n    return reporter.getClients();\n}",
    "new_code": "if (reporter != null && reporter.getClients() != null) {\n    return reporter.getClients();\n}"
  }
]
```

Rules:
- `old_code` must match the file **exactly** — whitespace, indentation, newlines. No fuzzy matches.
- Include 2–4 lines of surrounding context in `old_code` when the change is a single line, so replacement is unambiguous.
- Make every change **minimal**. If a one-line guard fixes the NPE, do not refactor the method.
- Prefer defensive guards + fail-soft over rewriting logic, unless the root cause really is broken logic.

## How to apply

Preferred — use the Edit tool directly with the old/new strings you identified. Only use string-replacement scripts if the Edit tool fails (e.g., non-unique match).

## Common fix patterns (IoTOps-specific)

- **Null guards in KTable value-joiners** — the join partner can arrive late; defensive null-check on the secondary stream's value.
- **Missing `Optional.ofNullable` wrap** around RocksDB `ReadOnlyKeyValueStore.get()` — returns `null` for keys evicted by TTL.
- **Stale Avro / Protobuf deserialiser cache** — fix is usually a `pom.xml` version bump in the schema dependency, not source code.
- **`TimeoutException` during broker reconnect** — almost always an infrastructure issue. Your RCA should have marked this `is_internal_issue: false`; do not try to patch Kafka consumer configs.
- **ClickHouse / ArangoDB write rejections** — usually schema drift. Fix is often in the entity or repository class that builds the write, not the client driver.

## After applying

```bash
git status           # confirm only expected files changed
git diff             # review the actual diff
git diff --stat      # confirm minimal footprint
git commit -am "fix(${AFFECTED_REPO}): ${ROOT_CAUSE_SHORT}"
```

Save the diff to `/tmp/patch.diff` so it can be included in the Verifier issue description.
