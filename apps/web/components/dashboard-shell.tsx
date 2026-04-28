"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Activity, Home, KeyRound, LockKeyhole, LogOut, Menu, PanelLeftClose, PanelLeftOpen, Shield, X } from "lucide-react";
import { useState, type ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "@/components/theme-toggle";
import { useRequireAuth } from "@/lib/auth-context";
import { cn } from "@/lib/utils";

const links = [
  { href: "/dashboard", label: "Dashboard", icon: Home },
  { href: "/workspaces", label: "Workspaces", icon: Shield },
  { href: "/secrets", label: "Secrets", icon: KeyRound },
  { href: "/audit", label: "Audit Log", icon: Activity }
];

export function DashboardShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const auth = useRequireAuth();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [collapsed, setCollapsed] = useState(false);

  if (!auth.ready || !auth.token) {
    return <main className="grid min-h-screen place-items-center bg-slate-50 text-sm text-slate-600 dark:bg-slate-950 dark:text-slate-300">Loading session...</main>;
  }

  const sidebar = (
    <aside className={cn("flex min-h-screen flex-col border-r border-slate-200 bg-slate-950 text-white transition-all dark:border-slate-800", collapsed ? "w-20" : "w-64")}>
      <div className="flex h-16 items-center justify-between px-4">
        <div className="flex items-center gap-2 font-semibold">
          <LockKeyhole className="h-5 w-5 text-teal-300" />
          {!collapsed && <span>DevsVault</span>}
        </div>
        <Button type="button" variant="ghost" size="icon" className="hidden text-white hover:bg-slate-800 lg:inline-flex" onClick={() => setCollapsed((value) => !value)} aria-label="Toggle sidebar">
          {collapsed ? <PanelLeftOpen className="h-4 w-4" /> : <PanelLeftClose className="h-4 w-4" />}
        </Button>
        <Button type="button" variant="ghost" size="icon" className="text-white hover:bg-slate-800 lg:hidden" onClick={() => setMobileOpen(false)} aria-label="Close menu">
          <X className="h-4 w-4" />
        </Button>
      </div>
      <nav className="grid gap-1 px-3 py-4">
        {links.map((item) => {
          const active = pathname === item.href || (item.href !== "/dashboard" && pathname.startsWith(item.href));
          const Icon = item.icon;
          return (
            <Link className={cn("flex items-center gap-3 rounded-md px-3 py-2 text-sm text-slate-300 hover:bg-slate-800 hover:text-white", active && "bg-slate-800 text-white")} href={item.href} key={item.href} onClick={() => setMobileOpen(false)}>
              <Icon className="h-4 w-4 shrink-0" />
              {!collapsed && <span>{item.label}</span>}
            </Link>
          );
        })}
      </nav>
    </aside>
  );

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-950">
      <div className="hidden lg:fixed lg:inset-y-0 lg:left-0 lg:block">{sidebar}</div>
      {mobileOpen && <div className="fixed inset-0 z-50 flex bg-slate-950/40 lg:hidden">{sidebar}</div>}
      <div className={cn("transition-all", collapsed ? "lg:pl-20" : "lg:pl-64")}>
        <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-slate-200 bg-white/95 px-4 backdrop-blur dark:border-slate-800 dark:bg-slate-950/95 sm:px-6">
          <div className="flex items-center gap-3">
            <Button type="button" variant="outline" size="icon" className="lg:hidden" onClick={() => setMobileOpen(true)} aria-label="Open menu">
              <Menu className="h-4 w-4" />
            </Button>
            <div className="min-w-0">
              <p className="truncate text-sm font-medium text-slate-900 dark:text-slate-100">{auth.subject}</p>
              <p className="text-xs text-slate-500 dark:text-slate-400">{auth.actorType} · {auth.apiUrl}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <ThemeToggle />
            <Button type="button" variant="outline" onClick={auth.logout}><LogOut className="h-4 w-4" />Logout</Button>
          </div>
        </header>
        <main className="mx-auto grid w-full max-w-7xl gap-6 p-4 sm:p-6">{children}</main>
      </div>
    </div>
  );
}