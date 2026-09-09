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
  apiKey: string | null;
  isLoading: boolean;
  setAccessToken: (token: string | null) => void;
  setUser: (user: User | null) => void;
  setApiKey: (apiKey: string | null) => void;
  login: (data: LoginFormData) => Promise<void>;
  register: (data: RegisterFormData) => Promise<RegisterResponse>;
  logout: () => Promise<void>;
}

interface LoginResponse {
  access_token: string;
  api_key?: string;
}

interface RegisterResponse {
  id: string;
  name: string;
  email: string;
  access_token: string;
  api_key: string;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  accessToken: null,
  apiKey: null,
  isLoading: false,

  setAccessToken: (token) => {
    set({ accessToken: token });
  },

  setUser: (user) => set({ user }),

  setApiKey: (apiKey) => set({ apiKey }),

  login: async (data) => {
    set({ isLoading: true });
    try {
      const { data: res } = await api.post<LoginResponse>("/v1/login", data);
      set({ accessToken: res.access_token });
      if (res.api_key) {
        set({ apiKey: res.api_key });
      }
    } finally {
      set({ isLoading: false });
    }
  },

  register: async (data) => {
    set({ isLoading: true });
    try {
      const { data: res } = await api.post<RegisterResponse>("/v1/users", {
        name: data.name,
        email: data.email,
        password: data.password,
      });
      set({
        user: { id: res.id, name: res.name, email: res.email },
        accessToken: res.access_token,
        apiKey: res.api_key,
      });
      return res;
    } finally {
      set({ isLoading: false });
    }
  },

  logout: async () => {
    try {
      await api.post("/v1/auth/logout");
    } finally {
      set({ user: null, accessToken: null, apiKey: null });
    }
  },
}));
