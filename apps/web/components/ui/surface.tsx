import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "@/lib/utils";

export function Card({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("rounded-lg border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900", className)} {...props} />;
}

export function Badge({ tone = "neutral", children }: { tone?: "neutral" | "success" | "danger" | "warning"; children: ReactNode }) {
  const tones = {
    neutral: "bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-200",
    success: "bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200",
    danger: "bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-200",
    warning: "bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-200"
  };
  return <span className={cn("inline-flex w-fit rounded-full px-2.5 py-1 text-xs font-medium", tones[tone])}>{children}</span>;
}

export function EmptyState({ title, action }: { title: string; action?: ReactNode }) {
  return (
    <div className="grid place-items-center gap-3 rounded-lg border border-dashed border-slate-300 p-8 text-center dark:border-slate-700">
      <p className="text-sm text-slate-600 dark:text-slate-300">{title}</p>
      {action}
    </div>
  );
}

export function SkeletonRows({ rows = 5, columns = 4 }: { rows?: number; columns?: number }) {
  return (
    <div className="grid gap-2">
      {Array.from({ length: rows }).map((_, row) => (
        <div className="grid gap-3 rounded-md border border-slate-200 p-3 dark:border-slate-800" style={{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` }} key={row}>
          {Array.from({ length: columns }).map((__, column) => <div className="h-4 animate-pulse rounded bg-slate-200 dark:bg-slate-800" key={column} />)}
        </div>
      ))}
    </div>
  );
}