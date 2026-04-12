import { Extension } from "@tiptap/core";

export function createSubmitExtension(onSubmit: () => void) {
  return Extension.create({
    name: "submitShortcut",
    addKeyboardShortcuts() {
      return {
        Enter: () => {
          onSubmit();
          return true;
        },
        "Shift-Enter": () => false,
        "Mod-Enter": () => {
          onSubmit();
          return true;
        },
      };
    },
  });
}
