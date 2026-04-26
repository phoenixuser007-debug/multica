"use client";

import { Monitor, Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { useEffect, useState, type ReactNode } from "react";
import { Button } from "@multica/ui/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@multica/ui/components/ui/dropdown-menu";
import { cn } from "@multica/ui/lib/utils";

const THEME_OPTIONS: { value: string; label: string; icon: ReactNode }[] = [
  { value: "light", label: "Light", icon: <Sun className="size-4" /> },
  { value: "dark", label: "Dark", icon: <Moon className="size-4" /> },
  { value: "system", label: "System", icon: <Monitor className="size-4" /> },
];

export function DocsSettings() {
  const { theme, setTheme } = useTheme();

  // Gate theme reads until mount — next-themes is SSR-incompatible and
  // would otherwise cause a hydration flash of the wrong icon.
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);

  const activeTheme = mounted ? (theme ?? "system") : "system";
  const activeThemeOption =
    THEME_OPTIONS.find((o) => o.value === activeTheme) ?? THEME_OPTIONS[2]!;

  return (
    <div className="flex w-full items-center justify-end gap-2">
      {/* Theme icon button. Matched height to the sm pill via
          the icon-sm size token; without this the icon variant defaults
          to 32px while size="sm" is 28px, misaligning them. */}
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button
              variant="ghost"
              size="icon-sm"
              className="shrink-0 text-muted-foreground"
              aria-label="Switch theme"
            >
              {activeThemeOption.icon}
            </Button>
          }
        />
        <DropdownMenuContent align="end" side="top" className="min-w-[140px]">
          {THEME_OPTIONS.map((opt) => (
            <DropdownMenuItem
              key={opt.value}
              onClick={() => setTheme(opt.value)}
              className={cn(
                "gap-2",
                opt.value === activeTheme && "bg-accent",
              )}
            >
              {opt.icon}
              {opt.label}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
