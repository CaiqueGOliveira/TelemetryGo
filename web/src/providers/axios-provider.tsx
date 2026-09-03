"use client";

import * as React from "react";
import type { AxiosInstance } from "axios";
import { createApiClient } from "@/lib/api";

const AxiosContext = React.createContext<AxiosInstance | null>(null);

export function AxiosProvider({ children }: { children: React.ReactNode }) {
  const [client] = React.useState<AxiosInstance>(() => createApiClient());

  return (
    <AxiosContext.Provider value={client}>{children}</AxiosContext.Provider>
  );
}

export function useAxios(): AxiosInstance {
  const context = React.useContext(AxiosContext);
  if (!context) {
    throw new Error("useAxios must be used within an AxiosProvider");
  }
  return context;
}
