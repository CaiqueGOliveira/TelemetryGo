"use client";

import { create } from "zustand";
import { api } from "@/lib/api";

export interface TelemetryEvent {
  id: string;
  type: string;
  service: string;
  message: string;
  severity: "info" | "warning" | "critical";
  timestamp: string;
}

const defaultEvents: TelemetryEvent[] = [
  { id: "1", type: "deploy", service: "api-gateway", message: "Deploy realizado com sucesso", severity: "info", timestamp: "2026-09-03T10:00:00Z" },
  { id: "2", type: "threshold", service: "worker", message: "Uso de memória acima do limite", severity: "warning", timestamp: "2026-09-03T10:05:00Z" },
  { id: "3", type: "error", service: "auth-service", message: "Erro ao autenticar usuário", severity: "critical", timestamp: "2026-09-03T10:10:00Z" },
];

interface EventsState {
  events: TelemetryEvent[];
  isLoading: boolean;
  error: string | null;
  loadEvents: () => Promise<void>;
}

export const useEventsStore = create<EventsState>((set) => ({
  events: defaultEvents,
  isLoading: false,
  error: null,

  loadEvents: async () => {
    set({ isLoading: true, error: null });
    try {
      const { data } = await api.get<TelemetryEvent[]>("/v1/events");
      set({ events: data });
    } catch {
      set({ events: defaultEvents });
    } finally {
      set({ isLoading: false });
    }
  },
}));
