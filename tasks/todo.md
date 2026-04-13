## 2026-04-12 Remove Legacy Locale Artifacts

- [x] Inventory non-English locale/docs artifacts and linked surfaces
- [x] Remove retired locale files, translated docs, and stale links
- [x] Collapse landing localization plumbing back to English-only
- [x] Run verification and confirm no retired locale artifacts remain

## Review

- Removed the retired localized README, landing locale file, locale cookie sync, and related links/imports.
- Collapsed landing locale plumbing to a single English dictionary and removed the footer language switcher plus locale-detection font/cookie logic.
- Removed non-English repository docs that were still checked in under `docs/`.
- Verification passed:
  - `pnpm typecheck`
  - `pnpm --filter @multica/web typecheck`
  - explicit identifier/file-name search returned no matches
  - targeted CJK scan across `README.md`, `apps/`, and `docs/` returned no matches

## 2026-04-12 Remove Retired Provider References

- [x] Inventory retired provider references across backend, UI, docs, and tests
- [x] Remove runtime/backend support and stale provider assets with minimal scope
- [x] Remove product/docs/landing references in English and retired localized copy
- [x] Run targeted verification and confirm no retired-provider references remain via repo-wide search

## Review

- Removed retired provider support from `server/pkg/agent` and daemon discovery/config paths, including deleting the retired backend implementations and their tests.
- Removed retired provider branding and copy from landing components, runtime logos, docs, README files, and localized landing changelog entries.
- A repo-wide search for the retired provider names now returns no matches.
- Verification passed:
  - `pnpm typecheck`
  - `pnpm --filter @multica/views typecheck`
  - `cd server && go test ./...`
- `make check` did not complete because `packages/views/auth/login-page.test.tsx` already fails in `@multica/views#test` with `TypeError: localStorage.clear is not a function`.

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
