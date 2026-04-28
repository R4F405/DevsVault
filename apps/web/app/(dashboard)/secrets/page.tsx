"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Copy, Eye, EyeOff, Plus, RotateCcw, ShieldX } from "lucide-react";
import { useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { FieldError, Input, Label, Select } from "@/components/ui/form";
import { Modal } from "@/components/ui/modal";
import { Card, EmptyState, SkeletonRows } from "@/components/ui/surface";
import { type SecretMetadata } from "@/lib/api-client";
import { useApiClient } from "@/lib/auth-context";
import { formatDate } from "@/lib/utils";

const secretSchema = z.object({ workspace_id: z.string().min(1), project_id: z.string().min(1), environment_id: z.string().min(1), name: z.string().min(1).max(120), value: z.string().min(1) });
const rotateSchema = z.object({ value: z.string().min(1) });
type SecretForm = z.infer<typeof secretSchema>;
type RotateForm = z.infer<typeof rotateSchema>;

export default function SecretsPage() {
  return <Suspense fallback={<main className="text-sm text-slate-600 dark:text-slate-300">Loading secrets...</main>}><SecretsContent /></Suspense>;
}

function SecretsContent() {
  const api = useApiClient();
  const params = useSearchParams();
  const queryClient = useQueryClient();
  const [workspaceId, setWorkspaceId] = useState(params.get("workspace") ?? "");
  const [projectId, setProjectId] = useState(params.get("project") ?? "");
  const [environmentId, setEnvironmentId] = useState(params.get("environment") ?? "");
  const [newOpen, setNewOpen] = useState(false);
  const [showNewValue, setShowNewValue] = useState(false);
  const [viewSecret, setViewSecret] = useState<SecretMetadata | null>(null);
  const [valueNonce, setValueNonce] = useState(0);
  const [rotateSecret, setRotateSecret] = useState<SecretMetadata | null>(null);
  const [revokeSecret, setRevokeSecret] = useState<SecretMetadata | null>(null);

  const workspaces = useQuery({ queryKey: ["workspaces"], queryFn: () => api.listWorkspaces() });
  const projects = useQuery({ queryKey: ["projects", workspaceId], queryFn: () => api.listProjects(workspaceId), enabled: workspaceId !== "" });
  const environments = useQuery({ queryKey: ["environments", projectId], queryFn: () => api.listEnvironments(projectId), enabled: projectId !== "" });
  const secrets = useQuery({ queryKey: ["secrets", workspaceId, projectId, environmentId], queryFn: () => api.listSecrets({ workspaceId, projectId, environmentId }) });
  const secretValue = useQuery({ queryKey: ["secret-value", viewSecret?.logical_path, valueNonce], queryFn: () => api.resolveSecret(viewSecret?.logical_path ?? ""), enabled: viewSecret !== null, staleTime: 0, gcTime: 0, retry: false });

  const filtered = secrets.data ?? [];

  const form = useForm<SecretForm>({ resolver: zodResolver(secretSchema), defaultValues: { workspace_id: workspaceId, project_id: projectId, environment_id: environmentId, name: "", value: "" } });
  const watchedWorkspace = form.watch("workspace_id");
  const watchedProject = form.watch("project_id");
  const formProjects = useQuery({ queryKey: ["projects", watchedWorkspace, "form"], queryFn: () => api.listProjects(watchedWorkspace), enabled: watchedWorkspace !== "" });
  const formEnvironments = useQuery({ queryKey: ["environments", watchedProject, "form"], queryFn: () => api.listEnvironments(watchedProject), enabled: watchedProject !== "" });
  const rotateForm = useForm<RotateForm>({ resolver: zodResolver(rotateSchema), defaultValues: { value: "" } });

  useEffect(() => { form.setValue("workspace_id", workspaceId); form.setValue("project_id", projectId); form.setValue("environment_id", environmentId); }, [environmentId, form, projectId, workspaceId]);

  const create = useMutation({ mutationFn: (values: SecretForm) => api.createSecret(values), onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["secrets"] }); form.reset({ workspace_id: workspaceId, project_id: projectId, environment_id: environmentId, name: "", value: "" }); setNewOpen(false); setShowNewValue(false); } });
  const rotate = useMutation({ mutationFn: (values: RotateForm) => api.rotateSecret(rotateSecret?.id ?? "", values.value), onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["secrets"] }); rotateForm.reset(); setRotateSecret(null); } });
  const revoke = useMutation({ mutationFn: () => api.revokeSecretVersion(revokeSecret?.id ?? "", revokeSecret?.active_version ?? 0), onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["secrets"] }); setRevokeSecret(null); } });

  function clearProject() { setProjectId(""); setEnvironmentId(""); }
  function closeValue() { setViewSecret(null); setValueNonce(0); }

  return (
    <section className="grid gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3"><div><h1 className="text-2xl font-semibold">Secrets</h1><p className="text-sm text-slate-600 dark:text-slate-300">Values are never listed and only resolved on explicit request.</p></div><Button onClick={() => setNewOpen(true)}><Plus className="h-4 w-4" />New Secret</Button></div>
      <Card className="grid gap-4">
        <div className="grid gap-3 md:grid-cols-3"><Label>Workspace<Select value={workspaceId} onChange={(event) => { setWorkspaceId(event.target.value); clearProject(); }}><option value="">All workspaces</option>{workspaces.data?.map((item) => <option value={item.id} key={item.id}>{item.name}</option>)}</Select></Label><Label>Project<Select value={projectId} onChange={(event) => { setProjectId(event.target.value); setEnvironmentId(""); }} disabled={!workspaceId}><option value="">All projects</option>{projects.data?.map((item) => <option value={item.id} key={item.id}>{item.name}</option>)}</Select></Label><Label>Environment<Select value={environmentId} onChange={(event) => setEnvironmentId(event.target.value)} disabled={!projectId}><option value="">All environments</option>{environments.data?.map((item) => <option value={item.id} key={item.id}>{item.name}</option>)}</Select></Label></div>
        {secrets.isLoading && <SkeletonRows rows={7} columns={4} />}
        {secrets.isError && <p className="text-sm text-red-700">Secrets could not be loaded.</p>}
        {secrets.data?.length === 0 && <EmptyState title="No secrets yet." action={<Button onClick={() => setNewOpen(true)}>New Secret</Button>} />}
        {!!secrets.data?.length && filtered.length === 0 && <EmptyState title="No secrets match the selected filters." />}
        {!!filtered.length && <div className="overflow-x-auto"><table className="w-full text-sm"><thead><tr className="border-b text-left text-slate-500 dark:border-slate-800"><th className="py-2">Path</th><th>Version</th><th>Updated</th><th className="text-right">Actions</th></tr></thead><tbody>{filtered.map((secret) => <tr className="border-b last:border-0 dark:border-slate-800" key={secret.id}><td className="py-3"><code className="rounded bg-slate-100 px-2 py-1 dark:bg-slate-800">{secret.logical_path}</code></td><td>v{secret.active_version}</td><td>{formatDate(secret.updated_at)}</td><td className="flex justify-end gap-2"><Button variant="outline" size="sm" onClick={() => { setViewSecret(secret); setValueNonce(Date.now()); }}><Eye className="h-4 w-4" />View</Button><Button variant="outline" size="sm" onClick={() => setRotateSecret(secret)}><RotateCcw className="h-4 w-4" />Rotate</Button><Button variant="outline" size="sm" onClick={() => setRevokeSecret(secret)}><ShieldX className="h-4 w-4" />Revoke</Button></td></tr>)}</tbody></table></div>}
      </Card>

      <Modal title="New Secret" open={newOpen} onClose={() => setNewOpen(false)} wide><form className="grid gap-4" onSubmit={form.handleSubmit((values) => create.mutate(values))}><div className="grid gap-3 md:grid-cols-3"><Label>Workspace<Select {...form.register("workspace_id")}><option value="">Select</option>{workspaces.data?.map((item) => <option value={item.id} key={item.id}>{item.name}</option>)}</Select><FieldError message={form.formState.errors.workspace_id?.message} /></Label><Label>Project<Select {...form.register("project_id")} disabled={!watchedWorkspace}><option value="">Select</option>{formProjects.data?.map((item) => <option value={item.id} key={item.id}>{item.name}</option>)}</Select><FieldError message={form.formState.errors.project_id?.message} /></Label><Label>Environment<Select {...form.register("environment_id")} disabled={!watchedProject}><option value="">Select</option>{formEnvironments.data?.map((item) => <option value={item.id} key={item.id}>{item.name}</option>)}</Select><FieldError message={form.formState.errors.environment_id?.message} /></Label></div><Label>Name<Input {...form.register("name")} /><FieldError message={form.formState.errors.name?.message} /></Label><Label>Value<div className="flex gap-2"><Input type={showNewValue ? "text" : "password"} {...form.register("value")} /><Button type="button" variant="outline" size="icon" onClick={() => setShowNewValue((value) => !value)} aria-label="Toggle value visibility">{showNewValue ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}</Button></div><FieldError message={form.formState.errors.value?.message} /></Label>{create.isError && <p className="text-sm text-red-700">Secret could not be created.</p>}<Button type="submit" disabled={create.isPending}>{create.isPending ? "Creating..." : "Create secret"}</Button></form></Modal>

      <Modal title="Secret Value" open={viewSecret !== null} onClose={closeValue}>{secretValue.isLoading && <p className="text-sm text-slate-600 dark:text-slate-300">Resolving value...</p>}{secretValue.isError && <p className="text-sm text-red-700">Secret value could not be resolved.</p>}{secretValue.data && <div className="grid gap-4"><div className="rounded-md border border-slate-200 bg-slate-50 p-3 font-mono text-sm dark:border-slate-800 dark:bg-slate-950">{secretValue.data.value}</div><Button type="button" variant="outline" onClick={() => void navigator.clipboard.writeText(secretValue.data.value)}><Copy className="h-4 w-4" />Copy</Button></div>}</Modal>

      <Modal title="Rotate Secret" open={rotateSecret !== null} onClose={() => setRotateSecret(null)}><form className="grid gap-4" onSubmit={rotateForm.handleSubmit((values) => rotate.mutate(values))}><Label>New value<Input type="password" {...rotateForm.register("value")} /><FieldError message={rotateForm.formState.errors.value?.message} /></Label>{rotate.isError && <p className="text-sm text-red-700">Secret could not be rotated.</p>}<Button type="submit" disabled={rotate.isPending}>{rotate.isPending ? "Rotating..." : "Rotate secret"}</Button></form></Modal>

      <Modal title="Revoke Version" open={revokeSecret !== null} onClose={() => setRevokeSecret(null)}><div className="grid gap-4"><p className="text-sm text-slate-600 dark:text-slate-300">Revoke active version v{revokeSecret?.active_version} for this secret?</p>{revoke.isError && <p className="text-sm text-red-700">Version could not be revoked.</p>}<Button type="button" variant="danger" disabled={revoke.isPending} onClick={() => revoke.mutate()}>{revoke.isPending ? "Revoking..." : "Confirm revoke"}</Button></div></Modal>
    </section>
  );
}