import { describe, expect, it } from "vitest";
import { preprocessLinks } from "@multica/ui/markdown/linkify";

// The bug: linkify-it does not treat full-width punctuation as a URL boundary,
// so the href can swallow trailing punctuation and later text. The fix
// truncates the detected URL at the first full-width punctuation character.

describe("preprocessLinks — full-width punctuation boundary", () => {
  it("stops URL at ideographic full stop 。", () => {
    const out = preprocessLinks("see https://example.com/path。then continue");
    expect(out).toBe("see [https://example.com/path](https://example.com/path)。then continue");
  });

  it("stops URL at fullwidth comma ，", () => {
    const out = preprocessLinks("open https://example.com/a，and more");
    expect(out).toBe("open [https://example.com/a](https://example.com/a)，and more");
  });

  it("stops URL at ideographic comma 、", () => {
    const out = preprocessLinks("two links https://a.com/x、https://b.com/y");
    expect(out).toBe(
      "two links [https://a.com/x](https://a.com/x)、[https://b.com/y](https://b.com/y)",
    );
  });

  it("stops URL at fullwidth right paren ）", () => {
    const out = preprocessLinks("（see https://example.com/x）after");
    expect(out).toBe("（see [https://example.com/x](https://example.com/x)）after");
  });

  it("stops URL at corner bracket 」", () => {
    const out = preprocessLinks("「https://example.com/a」after");
    expect(out).toBe("「[https://example.com/a](https://example.com/a)」after");
  });

  it("stops URL at fullwidth exclamation ！", () => {
    const out = preprocessLinks("great https://example.com/x！continue");
    expect(out).toBe("great [https://example.com/x](https://example.com/x)！continue");
  });

  it("handles the original bug report (PR link then 。 then more text)", () => {
    const out = preprocessLinks(
      "Merged PR #1623: https://github.com/multica-ai/multica/pull/1623。merge commit",
    );
    expect(out).toBe(
      "Merged PR #1623: [https://github.com/multica-ai/multica/pull/1623](https://github.com/multica-ai/multica/pull/1623)。merge commit",
    );
  });

  it("does not swallow the entire remainder when there is no trailing space", () => {
    const out = preprocessLinks("https://github.com/x/y/issues/1619。next");
    expect(out).toBe(
      "[https://github.com/x/y/issues/1619](https://github.com/x/y/issues/1619)。next",
    );
  });

  it("preserves ASCII trailing period handling (no regression)", () => {
    const out = preprocessLinks("visit https://example.com/path. next.");
    expect(out).toBe("visit [https://example.com/path](https://example.com/path). next.");
  });

  it("preserves plain URL with no trailing punctuation (no regression)", () => {
    const out = preprocessLinks("go https://example.com/path");
    expect(out).toBe("go [https://example.com/path](https://example.com/path)");
  });

  it("preserves non-ASCII letters inside URL path (only trims on punctuation)", () => {
    const out = preprocessLinks("https://example.com/wiki/かな note");
    expect(out).toBe(
      "[https://example.com/wiki/かな](https://example.com/wiki/かな) note",
    );
  });

  it("does not re-link an already-linked URL that contains 。", () => {
    // If a user or upstream already wrote [text](url。), we leave it alone.
    const input = "see [link](https://example.com/x。)after";
    expect(preprocessLinks(input)).toBe(input);
  });
});
