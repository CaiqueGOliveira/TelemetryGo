"use client";

import { create } from "zustand";
import { api } from "@/lib/api";
import type { Metric, MetricSeries } from "@/types";

const defaultMetrics: Metric[] = [
  { id: "1", name: "Latency", service: "auth-service", value: 42, unit: "ms", status: "ok", timestamp: "2026-09-03T10:00:00Z" },
  { id: "2", name: "CPU", service: "api-gateway", value: 38, unit: "%", status: "ok", timestamp: "2026-09-03T10:00:00Z" },
  { id: "3", name: "Memory", service: "worker", value: 1.2, unit: "GB", status: "warn", timestamp: "2026-09-03T10:00:00Z" },
  { id: "4", name: "Error Rate", service: "auth-service", value: 2.1, unit: "%", status: "crit", timestamp: "2026-09-03T10:00:00Z" },
  { id: "5", name: "Requests", service: "api-gateway", value: 154, unit: "/s", status: "ok", timestamp: "2026-09-03T10:00:00Z" },
];

const defaultSeries: MetricSeries[] = [
  {
    name: "Requests",
    service: "api-gateway",
    unit: "/s",
    data: [
      { timestamp: "2026-09-03T09:00:00Z", value: 120 },
      { timestamp: "2026-09-03T09:10:00Z", value: 135 },
      { timestamp: "2026-09-03T09:20:00Z", value: 154 },
    ],
  },
];

interface MetricsState {
  metrics: Metric[];
  series: MetricSeries[];
  isLoading: boolean;
  error: string | null;
  loadMetrics: () => Promise<void>;
}

export const useMetricsStore = create<MetricsState>((set) => ({
  metrics: defaultMetrics,
  series: defaultSeries,
  isLoading: false,
  error: null,

  loadMetrics: async () => {
    set({ isLoading: true, error: null });
    try {
      const { data } = await api.get<Metric[]>("/v1/metrics");
      set({ metrics: data });
    } catch {
      set({ metrics: defaultMetrics });
    } finally {
      set({ isLoading: false });
    }
  },
}));
