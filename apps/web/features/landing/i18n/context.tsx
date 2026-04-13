"use client";

import { createContext, useContext } from "react";
import { en } from "./en";
import type { LandingDict } from "./types";

type LocaleContextValue = {
  t: LandingDict;
};

const LocaleContext = createContext<LocaleContextValue | null>(null);

export function LocaleProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <LocaleContext.Provider value={{ t: en }}>
      {children}
    </LocaleContext.Provider>
  );
}

export function useLocale() {
  const ctx = useContext(LocaleContext);
  if (!ctx) throw new Error("useLocale must be used within LocaleProvider");
  return ctx;
}
