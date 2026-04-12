import { describe, expect, it, vi } from "vitest";
import { createSubmitExtension } from "./submit-shortcut";

function getShortcuts(onSubmit: () => void) {
  const extension = createSubmitExtension(onSubmit) as {
    config?: { addKeyboardShortcuts?: () => Record<string, () => boolean> };
  };
  return extension.config?.addKeyboardShortcuts?.() ?? {};
}

describe("createSubmitExtension", () => {
  it("submits on Enter", () => {
    const onSubmit = vi.fn();
    const shortcuts = getShortcuts(onSubmit);
    const onEnter = shortcuts.Enter;

    expect(onEnter).toBeTypeOf("function");
    expect(onEnter?.()).toBe(true);
    expect(onSubmit).toHaveBeenCalledTimes(1);
  });

  it("keeps Shift+Enter for newlines", () => {
    const onSubmit = vi.fn();
    const shortcuts = getShortcuts(onSubmit);
    const onShiftEnter = shortcuts["Shift-Enter"];

    expect(onShiftEnter).toBeTypeOf("function");
    expect(onShiftEnter?.()).toBe(false);
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("still submits on Mod+Enter", () => {
    const onSubmit = vi.fn();
    const shortcuts = getShortcuts(onSubmit);
    const onModEnter = shortcuts["Mod-Enter"];

    expect(onModEnter).toBeTypeOf("function");
    expect(onModEnter?.()).toBe(true);
    expect(onSubmit).toHaveBeenCalledTimes(1);
  });
});
