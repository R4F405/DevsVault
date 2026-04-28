"use client";

import { useQuery } from "@tanstack/react-query";
import { Activity, KeyRound, Shield } from "lucide-react";
import { Card, SkeletonRows } from "@/components/ui/surface";
import { useApiClient } from "@/lib/auth-context";
import { formatDate } from "@/lib/utils";

export default function DashboardPage() {
  const api = useApiClient();
  const workspaces = useQuery({ queryKey: ["workspaces"], queryFn: () => api.listWorkspaces() });
  const secrets = useQuery({ queryKey: ["secrets"], queryFn: () => api.listSecrets() });
  const audit = useQuery({ queryKey: ["audit-events"], queryFn: () => api.listAuditEvents() });
  const loading = workspaces.isLoading || secrets.isLoading || audit.isLoading;
  const latest = audit.data?.[0];

  return (
    <section className="grid gap-6">
      <div>
        <h1 className="text-2xl font-semibold text-slate-950 dark:text-slate-50">Dashboard</h1>
        <p className="text-sm text-slate-600 dark:text-slate-300">Operational summary for the active DevsVault API.</p>
      </div>
      {loading ? <SkeletonRows rows={3} columns={3} /> : (
        <div className="grid gap-4 md:grid-cols-3">
          <Card><Shield className="mb-3 h-5 w-5 text-teal-700" /><p className="text-sm text-slate-500">Total workspaces</p><strong className="text-3xl">{workspaces.data?.length ?? 0}</strong></Card>
          <Card><KeyRound className="mb-3 h-5 w-5 text-teal-700" /><p className="text-sm text-slate-500">Total secrets</p><strong className="text-3xl">{secrets.data?.length ?? 0}</strong></Card>
          <Card><Activity className="mb-3 h-5 w-5 text-teal-700" /><p className="text-sm text-slate-500">Latest audit event</p><strong className="block truncate text-lg">{latest ? latest.action : "No events"}</strong><span className="text-sm text-slate-500">{latest ? formatDate(latest.occurred_at) : "-"}</span></Card>
        </div>
      )}
      {(workspaces.isError || secrets.isError || audit.isError) && <p className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800 dark:border-red-900 dark:bg-red-950 dark:text-red-200">Some dashboard data could not be loaded.</p>}
    </section>
  );
}