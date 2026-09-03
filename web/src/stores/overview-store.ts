"use client";

import { create } from "zustand";
import { api } from "@/lib/api";
import type { OverviewCard } from "@/types";

const defaultCards: OverviewCard[] = [
  { id: "services", title: "Serviços", value: "12", change: "+2", trend: "up" },
  {
    id: "active-metrics",
    title: "Métricas ativas",
    value: "1.2k",
    change: "+8%",
    trend: "up",
  },
  { id: "avg-cpu", title: "CPU média", value: "38%", change: "-5%", trend: "down" },
  {
    id: "storage",
    title: "Armazenamento",
    value: "68%",
    change: "+3%",
    trend: "up",
  },
];

interface OverviewState {
  cards: OverviewCard[];
  isLoading: boolean;
  error: string | null;
  loadOverview: () => Promise<void>;
}

export const useOverviewStore = create<OverviewState>((set) => ({
  cards: defaultCards,
  isLoading: false,
  error: null,

  loadOverview: async () => {
    set({ isLoading: true, error: null });
    try {
      const { data } = await api.get<OverviewCard[]>("/v1/overview");
      set({ cards: data });
    } catch {
      set({ cards: defaultCards });
    } finally {
      set({ isLoading: false });
    }
  },
}));
