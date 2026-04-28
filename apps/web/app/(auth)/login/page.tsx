"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { LockKeyhole } from "lucide-react";
import { Suspense, useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { useRouter } from "next/navigation";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/surface";
import { FieldError, Input, Label, Select } from "@/components/ui/form";
import { useAuth, useLoginMessage } from "@/lib/auth-context";

const loginSchema = z.object({
  apiUrl: z.string().url("Enter a valid API URL"),
  subject: z.string().min(1, "Subject is required"),
  actorType: z.enum(["user", "service"])
});

type LoginForm = z.infer<typeof loginSchema>;

export default function LoginPage() {
  return <Suspense fallback={<main className="grid min-h-screen place-items-center bg-slate-50 text-sm text-slate-600 dark:bg-slate-950 dark:text-slate-300">Loading login...</main>}><LoginContent /></Suspense>;
}

function LoginContent() {
  const router = useRouter();
  const auth = useAuth();
  const message = useLoginMessage();
  const [error, setError] = useState<string | null>(null);
  const form = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: { apiUrl: auth.apiUrl, subject: "", actorType: "user" }
  });

  useEffect(() => {
    if (auth.ready && auth.token) {
      router.replace("/dashboard");
    }
  }, [auth.ready, auth.token, router]);

  async function onSubmit(values: LoginForm) {
    setError(null);
    try {
      await auth.login(values);
      router.replace("/dashboard");
    } catch {
      setError("Login failed. Check the API URL and subject.");
    }
  }

  return (
    <main className="grid min-h-screen place-items-center bg-slate-50 p-4 dark:bg-slate-950">
      <Card className="w-full max-w-md">
        <div className="mb-6 flex items-center gap-3">
          <div className="grid h-10 w-10 place-items-center rounded-md bg-teal-700 text-white"><LockKeyhole className="h-5 w-5" /></div>
          <div>
            <h1 className="text-xl font-semibold text-slate-950 dark:text-slate-50">DevsVault</h1>
            <p className="text-sm text-slate-600 dark:text-slate-300">Sign in to the secrets control plane.</p>
          </div>
        </div>
        {message && <p className="mb-4 rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800 dark:border-amber-900 dark:bg-amber-950 dark:text-amber-200">{message}</p>}
        {error && <p className="mb-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800 dark:border-red-900 dark:bg-red-950 dark:text-red-200">{error}</p>}
        <form className="grid gap-4" onSubmit={form.handleSubmit(onSubmit)}>
          <Label>API URL<Input {...form.register("apiUrl")} /><FieldError message={form.formState.errors.apiUrl?.message} /></Label>
          <Label>Subject<Input {...form.register("subject")} placeholder="admin@example.local" /><FieldError message={form.formState.errors.subject?.message} /></Label>
          <Label>Actor Type<Select {...form.register("actorType")}><option value="user">user</option><option value="service">service</option></Select></Label>
          <Button type="submit" disabled={form.formState.isSubmitting}>{form.formState.isSubmitting ? "Signing in..." : "Sign in"}</Button>
        </form>
      </Card>
    </main>
  );
}