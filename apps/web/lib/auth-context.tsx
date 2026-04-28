"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { ApiClient, type ActorType } from "@/lib/api-client";

const sessionKey = "devsvault.session";
const defaultApiUrl = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

type StoredSession = {
  token: string;
  subject: string;
  actorType: ActorType;
  apiUrl: string;
  expiresAt: number;
};

type AuthContextValue = {
  token: string | null;
  subject: string | null;
  actorType: ActorType | null;
  apiUrl: string;
  expiresAt: number | null;
  ready: boolean;
  login: (input: { apiUrl: string; subject: string; actorType: ActorType }) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [session, setSession] = useState<StoredSession | null>(null);
  const [ready, setReady] = useState(false);

  const clearSession = useCallback(() => {
    setSession(null);
    if (typeof window !== "undefined") {
      window.sessionStorage.removeItem(sessionKey);
    }
  }, []);

  useEffect(() => {
    const raw = window.sessionStorage.getItem(sessionKey);
    if (!raw) {
      setReady(true);
      return;
    }
    try {
      const parsed = JSON.parse(raw) as StoredSession;
      if (!parsed.token || parsed.expiresAt <= Date.now()) {
        window.sessionStorage.removeItem(sessionKey);
        if (!pathname.startsWith("/login")) {
          router.replace("/login?message=Session%20expired");
        }
      } else {
        setSession(parsed);
      }
    } catch {
      window.sessionStorage.removeItem(sessionKey);
    } finally {
      setReady(true);
    }
  }, [pathname, router]);

  useEffect(() => {
    if (!session) {
      return;
    }
    const delay = Math.max(session.expiresAt - Date.now(), 0);
    const timeout = window.setTimeout(() => {
      clearSession();
      router.replace("/login?message=Session%20expired");
    }, delay);
    return () => window.clearTimeout(timeout);
  }, [clearSession, router, session]);

  const login = useCallback(async (input: { apiUrl: string; subject: string; actorType: ActorType }) => {
    const apiUrl = input.apiUrl.replace(/\/$/, "") || defaultApiUrl;
    const client = new ApiClient(apiUrl);
    const response = await client.login({ subject: input.subject, actor_type: input.actorType });
    const nextSession: StoredSession = {
      token: response.access_token,
      subject: input.subject,
      actorType: input.actorType,
      apiUrl,
      expiresAt: Date.now() + response.expires_in * 1000
    };
    setSession(nextSession);
    window.sessionStorage.setItem(sessionKey, JSON.stringify(nextSession));
  }, []);

  const logout = useCallback(() => {
    clearSession();
    router.replace("/login");
  }, [clearSession, router]);

  const value = useMemo<AuthContextValue>(() => ({
    token: session?.token ?? null,
    subject: session?.subject ?? null,
    actorType: session?.actorType ?? null,
    apiUrl: session?.apiUrl ?? defaultApiUrl,
    expiresAt: session?.expiresAt ?? null,
    ready,
    login,
    logout
  }), [login, logout, ready, session]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used inside AuthProvider");
  }
  return context;
}

export function useApiClient() {
  const { apiUrl, token } = useAuth();
  return useMemo(() => new ApiClient(apiUrl, token ?? undefined), [apiUrl, token]);
}

export function useRequireAuth() {
  const auth = useAuth();
  const router = useRouter();
  useEffect(() => {
    if (auth.ready && !auth.token) {
      router.replace("/login");
    }
  }, [auth.ready, auth.token, router]);
  return auth;
}

export function useLoginMessage() {
  const params = useSearchParams();
  return params.get("message");
}