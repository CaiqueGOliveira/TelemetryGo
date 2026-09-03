"use client";

import { create } from "zustand";
import { api } from "@/lib/api";
import type { LoginFormData, RegisterFormData } from "@/schemas/auth-schema";

interface User {
  id: string;
  name: string;
  email: string;
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  isLoading: boolean;
  setAccessToken: (token: string | null) => void;
  setUser: (user: User | null) => void;
  login: (data: LoginFormData) => Promise<void>;
  register: (data: RegisterFormData) => Promise<void>;
  logout: () => void;
}

function getStoredToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("access_token");
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  accessToken: getStoredToken(),
  isLoading: false,

  setAccessToken: (token) => {
    if (typeof window !== "undefined") {
      if (token) {
        localStorage.setItem("access_token", token);
      } else {
        localStorage.removeItem("access_token");
      }
    }
    set({ accessToken: token });
  },

  setUser: (user) => set({ user }),

  login: async (data) => {
    set({ isLoading: true });
    try {
      const { data: res } = await api.post<{ access_token: string }>("/v1/login", data);
      set({ accessToken: res.access_token });
      if (typeof window !== "undefined") {
        localStorage.setItem("access_token", res.access_token);
      }
    } finally {
      set({ isLoading: false });
    }
  },

  register: async (data) => {
    set({ isLoading: true });
    try {
      await api.post("/v1/users", {
        name: data.name,
        email: data.email,
        password: data.password,
      });
    } finally {
      set({ isLoading: false });
    }
  },

  logout: () => {
    if (typeof window !== "undefined") {
      localStorage.removeItem("access_token");
    }
    set({ user: null, accessToken: null });
  },
}));
