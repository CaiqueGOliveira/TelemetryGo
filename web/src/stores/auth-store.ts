"use client";

import { create } from "zustand";
import { api } from "@/lib/api";
import { clearSession, loadSession, saveSession } from "@/lib/session";
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
  isHydrated: boolean;
  setAccessToken: (token: string | null) => void;
  setUser: (user: User | null) => void;
  setApiKey: (apiKey: string | null) => void;
  fetchUser: () => Promise<void>;
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

interface ApiError {
  response?: {
    status: number;
  };
}

function isUnauthorized(error: unknown): boolean {
  return (
    typeof error === "object" &&
    error !== null &&
    (error as ApiError).response?.status === 401
  );
}

const initialSession = loadSession();

export const useAuthStore = create<AuthState>((set, get) => ({
  user: initialSession.user,
  accessToken: initialSession.accessToken,
  apiKey: initialSession.apiKey,
  isLoading: false,
  isHydrated: initialSession.accessToken !== null,

  setAccessToken: (token) => {
    set({ accessToken: token });
  },

  setUser: (user) => set({ user }),

  setApiKey: (apiKey) => set({ apiKey }),

  fetchUser: async () => {
    if (!get().accessToken) return;
    try {
      const { data } = await api.get<User>("/v1/me");
      set({ user: data, isHydrated: true });
    } catch (error) {
      if (isUnauthorized(error)) {
        set({ user: null, accessToken: null, apiKey: null, isHydrated: true });
        clearSession();
      }
    }
  },

  login: async (data) => {
    set({ isLoading: true });
    try {
      const { data: res } = await api.post<LoginResponse>("/v1/login", data);
      set({ accessToken: res.access_token, isHydrated: true });
      if (res.api_key) {
        useAuthStore.getState().setApiKey(res.api_key);
      }
    } finally {
      set({ isLoading: false });
      saveSession(get());
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
        isHydrated: true,
      });
      saveSession(get());
      return res;
    } finally {
      set({ isLoading: false });
    }
  },

  logout: async () => {
    try {
      await api.post("/v1/auth/logout");
    } finally {
      set({ user: null, accessToken: null, apiKey: null, isHydrated: true });
      saveSession(get());
    }
  },
}));