"use client";

import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Badge, Card, EmptyState, SkeletonRows } from "@/components/ui/surface";
import { Button } from "@/components/ui/button";
import { useApiClient } from "@/lib/auth-context";
import { formatDate } from "@/lib/utils";

const pageSize = 20;

export default function AuditPage() {
  const api = useApiClient();
  const [page, setPage] = useState(0);
  const query = useQuery({ queryKey: ["audit-events"], queryFn: () => api.listAuditEvents() });
  const events = useMemo(() => query.data ?? [], [query.data]);
  const pages = Math.max(Math.ceil(events.length / pageSize), 1);
  const pageItems = useMemo(() => events.slice(page * pageSize, page * pageSize + pageSize), [events, page]);

  return (
    <section className="grid gap-6">
      <div><h1 className="text-2xl font-semibold">Audit Log</h1><p className="text-sm text-slate-600 dark:text-slate-300">Latest 100 events returned by the API.</p></div>
      <Card>
        {query.isLoading && <SkeletonRows rows={8} columns={5} />}
        {query.isError && <p className="text-sm text-red-700">Audit events could not be loaded.</p>}
        {query.data?.length === 0 && <EmptyState title="No audit events returned by the API." />}
        {!!pageItems.length && <div className="overflow-x-auto"><table className="w-full text-sm"><thead><tr className="border-b text-left text-slate-500 dark:border-slate-800"><th className="py-2">Time</th><th>Actor</th><th>Action</th><th>Resource</th><th>Outcome</th></tr></thead><tbody>{pageItems.map((event) => <tr className="border-b last:border-0 dark:border-slate-800" key={event.id}><td className="py-3 whitespace-nowrap">{formatDate(event.occurred_at)}</td><td>{event.actor_id || event.actor_type}</td><td>{event.action}</td><td>{event.resource_type}{event.resource_id ? `/${event.resource_id}` : ""}</td><td><Badge tone={event.outcome === "success" ? "success" : event.outcome === "denied" ? "danger" : "warning"}>{event.outcome}</Badge></td></tr>)}</tbody></table></div>}
        {!!pageItems.length && <div className="mt-4 flex items-center justify-end gap-2"><Button variant="outline" size="sm" disabled={page === 0} onClick={() => setPage((value) => Math.max(value - 1, 0))}>Previous</Button><span className="text-sm text-slate-500">Page {page + 1} of {pages}</span><Button variant="outline" size="sm" disabled={page + 1 >= pages} onClick={() => setPage((value) => Math.min(value + 1, pages - 1))}>Next</Button></div>}
      </Card>
    </section>
  );
}