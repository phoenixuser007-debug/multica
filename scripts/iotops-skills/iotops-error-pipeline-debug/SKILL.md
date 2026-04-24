---
name: iotops-error-pipeline-debug
description: "Reference bundle of the 8 original stage SKILL.md files from the legacy iotops-error-monitor Python pipeline. Engineer-facing; attach to whoever maintains the monitor. Not attached to the runtime Scout/Fixer/Verifier agents."
allowed-tools: ["Read", "Grep", "Glob", "Edit", "Bash"]
---

# IoTOps Error Pipeline Debug Reference

This skill bundles the 8 stage-specific SKILL.md files from the legacy
iotops-error-monitor Python pipeline as reference material. They document
the historical implementation of the pipeline before it was replaced by
the Scout / Fixer / Verifier agent chain that now lives in multica.

## Bundled files

- `stage1-humio-filter/SKILL.md` — Humio LSQL filter and dedup store
- `stage2-log-context/SKILL.md` — ±30-min pod-context window and Humio deep-link builder
- `stage3-rca/SKILL.md` — LLM root-cause analysis (Copilot Models API, gpt-4o)
- `stage4-jira/SKILL.md` — JIRA CNX Bug creation via REST v2
- `stage5-fix-generation/SKILL.md` — LLM-generated unified-diff patch
- `stage6-devenv-validation/SKILL.md` — compile + lint inside dev-env-dev-1
- `stage7-pull-request/SKILL.md` — Stash draft PR creation flow
- `stage8-slack/SKILL.md` — Slack notification templates

## When to use

Attach to a human maintainer or engineer-support agent investigating how the
original Python pipeline worked — when porting a missing edge case into the new
Scout/Fixer/Verifier skills, or reviewing historical behaviour the current
chain does not yet implement.

Not attached to the runtime agents by default: their authoritative instructions
live in `iotops-scout`, `iotops-fix-apply`, and `iotops-verify-pr-slack`.
