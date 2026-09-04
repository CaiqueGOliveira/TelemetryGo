import axios, { type AxiosInstance } from "axios";
import { useAuthStore } from "@/stores/auth-store";

export function createApiClient(): AxiosInstance {
  const instance = axios.create({
    baseURL: "/api",
    headers: {
      "Content-Type": "application/json",
    },
    withCredentials: true,
  });

  instance.interceptors.request.use((config) => {
    const token = useAuthStore.getState().accessToken;
    if (token && !config.headers.Authorization) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  });

  instance.interceptors.response.use(
    (response) => response,
    async (error) => {
      const original = error.config;

      if (error.response?.status === 401 && !original._retry) {
        original._retry = true;
        try {
          const { data } = await axios.post<{ access_token: string }>(
            "/api/auth/refresh"
          );
          useAuthStore.getState().setAccessToken(data.access_token);
          original.headers.Authorization = `Bearer ${data.access_token}`;
          return instance(original);
        } catch {
          useAuthStore.getState().setAccessToken(null);
          useAuthStore.getState().setUser(null);
        }
      }

      return Promise.reject(error);
    }
  );

  return instance;
}

export const api = createApiClient();
