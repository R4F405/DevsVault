"use client";

import Link from "next/link";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { FieldError, Input, Label, Textarea } from "@/components/ui/form";
import { Modal } from "@/components/ui/modal";
import { Card, EmptyState, SkeletonRows } from "@/components/ui/surface";
import { type Workspace } from "@/lib/api-client";
import { useApiClient } from "@/lib/auth-context";
import { formatDate } from "@/lib/utils";

const schema = z.object({ name: z.string().min(1).max(100), slug: z.string().regex(/^[a-z0-9](?:[a-z0-9-]*[a-z0-9])$/).min(2).max(50), description: z.string().max(300).optional() });
type FormValues = z.infer<typeof schema>;

export default function WorkspacesPage() {
  const api = useApiClient();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Workspace | null>(null);
  const [deleting, setDeleting] = useState<Workspace | null>(null);
  const query = useQuery({ queryKey: ["workspaces"], queryFn: () => api.listWorkspaces() });
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { name: "", slug: "", description: "" } });
  const editForm = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { name: "", slug: "", description: "" } });
  const create = useMutation({ mutationFn: (values: FormValues) => api.createWorkspace({ name: values.name, slug: values.slug, description: values.description ?? "" }), onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["workspaces"] }); form.reset(); setOpen(false); } });
  const update = useMutation({ mutationFn: (values: FormValues) => api.updateWorkspace(editing?.id ?? "", { name: values.name, description: values.description ?? "" }), onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["workspaces"] }); setEditing(null); } });
  const remove = useMutation({ mutationFn: () => api.deleteWorkspace(deleting?.id ?? ""), onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["workspaces"] }); setDeleting(null); } });

  function openEdit(workspace: Workspace) {
    editForm.reset({ name: workspace.name, slug: workspace.slug, description: workspace.description ?? "" });
    setEditing(workspace);
  }

  return (
    <section className="grid gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3"><div><h1 className="text-2xl font-semibold">Workspaces</h1><p className="text-sm text-slate-600 dark:text-slate-300">Top-level tenancy boundaries for projects and secrets.</p></div><Button onClick={() => setOpen(true)}><Plus className="h-4 w-4" />New Workspace</Button></div>
      <Card>
        {query.isLoading && <SkeletonRows rows={5} columns={4} />}
        {query.isError && <p className="text-sm text-red-700">Workspaces could not be loaded.</p>}
        {query.data?.length === 0 && <EmptyState title="No workspaces yet." action={<Button onClick={() => setOpen(true)}>New Workspace</Button>} />}
        {!!query.data?.length && <div className="overflow-x-auto"><table className="w-full text-sm"><thead><tr className="border-b text-left text-slate-500 dark:border-slate-800"><th className="py-2">Name</th><th>Slug</th><th>Created</th><th className="text-right">Actions</th></tr></thead><tbody>{query.data.map((workspace) => <tr className="border-b last:border-0 dark:border-slate-800" key={workspace.id}><td className="py-3 font-medium"><Link className="text-teal-700 hover:underline dark:text-teal-300" href={`/workspaces/${workspace.id}/projects`}>{workspace.name}</Link></td><td>{workspace.slug}</td><td>{formatDate(workspace.created_at)}</td><td className="flex justify-end gap-2"><Button type="button" variant="outline" size="sm" onClick={() => openEdit(workspace)}>Edit</Button><Button type="button" variant="danger" size="sm" onClick={() => setDeleting(workspace)}>Delete</Button></td></tr>)}</tbody></table></div>}
      </Card>
      <Modal title="New Workspace" open={open} onClose={() => setOpen(false)}>
        <form className="grid gap-4" onSubmit={form.handleSubmit((values) => create.mutate(values))}>
          <Label>Name<Input {...form.register("name")} /><FieldError message={form.formState.errors.name?.message} /></Label>
          <Label>Slug<Input {...form.register("slug")} /><FieldError message={form.formState.errors.slug?.message} /></Label>
          <Label>Description<Textarea {...form.register("description")} /><FieldError message={form.formState.errors.description?.message} /></Label>
          {create.isError && <p className="text-sm text-red-700">Workspace could not be created.</p>}
          <Button type="submit" disabled={create.isPending}>{create.isPending ? "Creating..." : "Create workspace"}</Button>
        </form>
      </Modal>
      <Modal title="Edit Workspace" open={editing !== null} onClose={() => setEditing(null)}>
        <form className="grid gap-4" onSubmit={editForm.handleSubmit((values) => update.mutate(values))}>
          <Label>Name<Input {...editForm.register("name")} /><FieldError message={editForm.formState.errors.name?.message} /></Label>
          <Label>Slug<Input readOnly {...editForm.register("slug")} /><FieldError message={editForm.formState.errors.slug?.message} /></Label>
          <Label>Description<Textarea {...editForm.register("description")} /><FieldError message={editForm.formState.errors.description?.message} /></Label>
          {update.isError && <p className="text-sm text-red-700">Workspace could not be updated.</p>}
          <Button type="submit" disabled={update.isPending}>{update.isPending ? "Saving..." : "Save workspace"}</Button>
        </form>
      </Modal>
      <Modal title="Delete Workspace" open={deleting !== null} onClose={() => setDeleting(null)}>
        <div className="grid gap-4">
          <p className="text-sm text-slate-600 dark:text-slate-300">Delete workspace {deleting?.name}?</p>
          {remove.isError && <p className="text-sm text-red-700">Workspace could not be deleted.</p>}
          <Button type="button" variant="danger" disabled={remove.isPending} onClick={() => remove.mutate()}>{remove.isPending ? "Deleting..." : "Confirm delete"}</Button>
        </div>
      </Modal>
    </section>
  );
}