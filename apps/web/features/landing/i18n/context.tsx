"use client";

import { createContext, useContext, useMemo } from "react";
import { useConfigStore } from "@multica/core/config";
import { createEnDict } from "./en";
import type { LandingDict, Locale } from "./types";

const dictionaryFactories: Record<Locale, (allowSignup: boolean) => LandingDict> = {
  en: createEnDict,
};

type LocaleContextValue = {
  locale: Locale;
  t: LandingDict;
};

const LocaleContext = createContext<LocaleContextValue | null>(null);

export function LocaleProvider({
  children,
  initialLocale = "en",
}: {
  children: React.ReactNode;
  initialLocale?: Locale;
}) {
  const locale = initialLocale;
  const allowSignup = useConfigStore((state) => state.allowSignup);
  const t = useMemo(
    () => dictionaryFactories[locale](allowSignup),
    [allowSignup, locale],
  );

  return (
    <LocaleContext.Provider
      value={{ locale, t }}
    >
      {children}
    </LocaleContext.Provider>
  );
}

export function useLocale() {
  const ctx = useContext(LocaleContext);
  if (!ctx) throw new Error("useLocale must be used within LocaleProvider");
  return ctx;
}
