# Comment Enter Shortcut Bug

- [x] Inspect comment composer and current keyboard handling
- [x] Add a failing test for Enter and Shift+Enter submit/newline behavior
- [x] Implement the input handler fix with minimal scope
- [x] Run targeted verification and record results

## Review

- Root cause was shared editor submit behavior, not the comment form wrapper. `createSubmitExtension` only handled `Mod-Enter`, so comments and replies inserted a newline on plain Enter instead of submitting.
- Added regression coverage in `packages/views/editor/extensions/submit-shortcut.test.ts` for `Enter` submit, `Shift+Enter` newline, and `Mod+Enter` submit.
- Updated `packages/views/editor/extensions/submit-shortcut.ts` to submit on `Enter`, preserve multiline entry on `Shift+Enter`, and keep `Mod+Enter` submit.
- Fixed a separate package typecheck regression in `packages/views/runtimes/components/provider-logo.tsx` by replacing the invalid `lucide-react` GitHub import with an inline SVG logo component.
- Verification passed:
  - `pnpm --filter @multica/views exec vitest run editor/extensions/submit-shortcut.test.ts`
  - `pnpm --filter @multica/views typecheck`
