"use client";

import Link from "next/link";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { use, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { FieldError, Input, Label } from "@/components/ui/form";
import { Modal } from "@/components/ui/modal";
import { Card, EmptyState, SkeletonRows } from "@/components/ui/surface";
import { useApiClient } from "@/lib/auth-context";
import { formatDate } from "@/lib/utils";

const schema = z.object({ name: z.string().min(1).max(100), slug: z.string().regex(/^[a-z0-9](?:[a-z0-9-]*[a-z0-9])$/).min(2).max(50) });
type FormValues = z.infer<typeof schema>;

export default function EnvironmentsPage({ params }: { params: Promise<{ id: string; projectId: string }> }) {
  const { id, projectId } = use(params);
  const api = useApiClient();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const workspace = useQuery({ queryKey: ["workspace", id], queryFn: () => api.getWorkspace(id) });
  const project = useQuery({ queryKey: ["project", id, projectId], queryFn: () => api.getProject(id, projectId) });
  const environments = useQuery({ queryKey: ["environments", projectId], queryFn: () => api.listEnvironments(projectId) });
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { name: "", slug: "" } });
  const create = useMutation({ mutationFn: (values: FormValues) => api.createEnvironment(projectId, values), onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["environments", projectId] }); form.reset(); setOpen(false); } });

  return (
    <section className="grid gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3"><div><p className="text-sm text-slate-500"><Link href="/workspaces" className="hover:underline">Workspaces</Link> &gt; <Link href={`/workspaces/${id}/projects`} className="hover:underline">{workspace.data?.name ?? "Workspace"}</Link> &gt; {project.data?.name ?? "Project"}</p><h1 className="text-2xl font-semibold">Environments</h1></div><Button onClick={() => setOpen(true)}><Plus className="h-4 w-4" />New Environment</Button></div>
      <Card>
        {(workspace.isLoading || project.isLoading || environments.isLoading) && <SkeletonRows rows={5} columns={3} />}
        {(workspace.isError || project.isError || environments.isError) && <p className="text-sm text-red-700">Environments could not be loaded.</p>}
        {environments.data?.length === 0 && <EmptyState title="No environments yet." action={<Button onClick={() => setOpen(true)}>New Environment</Button>} />}
        {!!environments.data?.length && <div className="overflow-x-auto"><table className="w-full text-sm"><thead><tr className="border-b text-left text-slate-500 dark:border-slate-800"><th className="py-2">Name</th><th>Slug</th><th>Created</th><th></th></tr></thead><tbody>{environments.data.map((environment) => <tr className="border-b last:border-0 dark:border-slate-800" key={environment.id}><td className="py-3 font-medium">{environment.name}</td><td>{environment.slug}</td><td>{formatDate(environment.created_at)}</td><td className="text-right"><Button asChild variant="outline" size="sm"><Link href={`/secrets?workspace=${id}&project=${projectId}&environment=${environment.id}`}>View secrets</Link></Button></td></tr>)}</tbody></table></div>}
      </Card>
      <Modal title="New Environment" open={open} onClose={() => setOpen(false)}><form className="grid gap-4" onSubmit={form.handleSubmit((values) => create.mutate(values))}><Label>Name<Input {...form.register("name")} /><FieldError message={form.formState.errors.name?.message} /></Label><Label>Slug<Input {...form.register("slug")} /><FieldError message={form.formState.errors.slug?.message} /></Label>{create.isError && <p className="text-sm text-red-700">Environment could not be created.</p>}<Button type="submit" disabled={create.isPending}>{create.isPending ? "Creating..." : "Create environment"}</Button></form></Modal>
    </section>
  );
}