"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";

export default function Home() {
  const router = useRouter();
  const { ready, token } = useAuth();

  useEffect(() => {
    if (ready) {
      router.replace(token ? "/dashboard" : "/login");
    }
  }, [ready, router, token]);

  return <main className="grid min-h-screen place-items-center bg-slate-50 text-sm text-slate-600 dark:bg-slate-950 dark:text-slate-300">Loading DevsVault...</main>;
}