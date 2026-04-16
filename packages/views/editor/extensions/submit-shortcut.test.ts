import { describe, expect, it, vi } from "vitest";
import { createSubmitExtension } from "./submit-shortcut";

function getShortcuts(onSubmit: () => boolean, submitOnEnter = true) {
  const extension = createSubmitExtension(onSubmit, { submitOnEnter }) as {
    config?: { addKeyboardShortcuts?: () => Record<string, () => boolean> };
  };
  return extension.config?.addKeyboardShortcuts?.() ?? {};
}

describe("createSubmitExtension", () => {
  it("submits on Enter when submitOnEnter=true", () => {
    const onSubmit = vi.fn().mockReturnValue(true);
    const shortcuts = getShortcuts(onSubmit, true);
    const onEnter = shortcuts.Enter;

    expect(onEnter).toBeTypeOf("function");
    expect(onEnter?.()).toBe(true);
    expect(onSubmit).toHaveBeenCalledTimes(1);
  });

  it("does not register Enter when submitOnEnter=false", () => {
    const onSubmit = vi.fn().mockReturnValue(true);
    const shortcuts = getShortcuts(onSubmit, false);

    expect(shortcuts.Enter).toBeUndefined();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("still submits on Mod+Enter", () => {
    const onSubmit = vi.fn().mockReturnValue(true);
    const shortcuts = getShortcuts(onSubmit);
    const onModEnter = shortcuts["Mod-Enter"];

    expect(onModEnter).toBeTypeOf("function");
    expect(onModEnter?.()).toBe(true);
    expect(onSubmit).toHaveBeenCalledTimes(1);
  });
});
