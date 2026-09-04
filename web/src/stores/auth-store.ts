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
  logout: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  accessToken: null,
  isLoading: false,

  setAccessToken: (token) => {
    set({ accessToken: token });
  },

  setUser: (user) => set({ user }),

  login: async (data) => {
    set({ isLoading: true });
    try {
      const { data: res } = await api.post<{ access_token: string }>("/v1/login", data);
      set({ accessToken: res.access_token });
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

  logout: async () => {
    try {
      await api.post("/auth/logout");
    } finally {
      set({ user: null, accessToken: null });
    }
  },
}));
