const STORAGE_KEY = "telemetrygo.auth";
const AUTH_COOKIE = "tg_session";

interface SessionData {
  user: {
    id: string;
    name: string;
    email: string;
  } | null;
  accessToken: string | null;
  apiKey: string | null;
}

function setAuthCookie() {
  if (typeof document === "undefined") return;
  document.cookie = `${AUTH_COOKIE}=1; path=/; max-age=604800; samesite=lax`;
}

function clearAuthCookie() {
  if (typeof document === "undefined") return;
  document.cookie = `${AUTH_COOKIE}=; path=/; max-age=0; samesite=lax`;
}

export function loadSession(): SessionData {
  if (typeof window === "undefined") return { user: null, accessToken: null, apiKey: null };
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return { user: null, accessToken: null, apiKey: null };
    const parsed = JSON.parse(raw) as SessionData;
    return {
      user: parsed.user ?? null,
      accessToken: parsed.accessToken ?? null,
      apiKey: parsed.apiKey ?? null,
    };
  } catch {
    return { user: null, accessToken: null, apiKey: null };
  }
}

export function saveSession(session: SessionData) {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
  } catch {}
  if (session.accessToken) {
    setAuthCookie();
  } else {
    clearAuthCookie();
  }
}

export function clearSession() {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.removeItem(STORAGE_KEY);
  } catch {}
  clearAuthCookie();
}