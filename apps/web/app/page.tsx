import { Activity, Braces, EyeOff, KeyRound, LockKeyhole, RotateCcw, ShieldCheck, Terminal, UsersRound } from "lucide-react";

const secrets = [
  { name: "DATABASE_URL", path: "acme/api/prod/DATABASE_URL", version: 7, rotation: "12d", access: "runtime-api", risk: "healthy" },
  { name: "STRIPE_TOKEN", path: "acme/billing/prod/STRIPE_TOKEN", version: 3, rotation: "89d", access: "billing-worker", risk: "review" },
  { name: "OIDC_CLIENT_SECRET", path: "acme/web/staging/OIDC_CLIENT_SECRET", version: 2, rotation: "34d", access: "web-admin", risk: "healthy" }
];

const audit = [
  { action: "secret.read", actor: "svc-runtime-api", target: "DATABASE_URL", outcome: "success", time: "2 min" },
  { action: "secret.rotate", actor: "admin@example.local", target: "STRIPE_TOKEN", outcome: "success", time: "18 min" },
  { action: "auth.authorize", actor: "dev@example.local", target: "OIDC_CLIENT_SECRET", outcome: "denied", time: "41 min" }
];

export default function Home() {
  return (
    <main className="shell">
      <aside className="sidebar" aria-label="Primary navigation">
        <div className="brand"><LockKeyhole size={22} /> DevsVault</div>
        <nav>
          <a className="active" href="#secrets"><KeyRound size={18} /> Secrets</a>
          <a href="#access"><UsersRound size={18} /> Access</a>
          <a href="#audit"><Activity size={18} /> Audit</a>
          <a href="#cli"><Terminal size={18} /> CLI</a>
        </nav>
      </aside>

      <section className="workspace">
        <header className="topbar">
          <div>
            <p className="eyebrow">workspace / acme</p>
            <h1>Secrets control plane</h1>
          </div>
          <button className="primary"><RotateCcw size={17} /> Rotate selected</button>
        </header>

        <section className="metrics" aria-label="Security summary">
          <div><ShieldCheck size={20} /><span>42 protected secrets</span><strong>0 plaintext</strong></div>
          <div><EyeOff size={20} /><span>Value exposure</span><strong>Explicit only</strong></div>
          <div><Braces size={20} /><span>Runtime paths</span><strong>18 active</strong></div>
        </section>

        <section id="secrets" className="panel">
          <div className="panelHeader">
            <div>
              <h2>Secret metadata</h2>
              <p>Values are hidden in listings and require read-value permission.</p>
            </div>
            <button className="secondary"><KeyRound size={16} /> New secret</button>
          </div>
          <div className="table" role="table" aria-label="Secrets metadata">
            <div className="row head" role="row">
              <span>Name</span><span>Path</span><span>Version</span><span>Rotation age</span><span>Access</span><span>Status</span>
            </div>
            {secrets.map((secret) => (
              <div className="row" role="row" key={secret.path}>
                <strong>{secret.name}</strong>
                <code>{secret.path}</code>
                <span>v{secret.version}</span>
                <span>{secret.rotation}</span>
                <span>{secret.access}</span>
                <span className={secret.risk === "review" ? "badge warn" : "badge ok"}>{secret.risk}</span>
              </div>
            ))}
          </div>
        </section>

        <section className="split">
          <div id="access" className="panel compact">
            <h2>Access posture</h2>
            <div className="accessItem"><span>admin</span><strong>7 permissions</strong></div>
            <div className="accessItem"><span>developer</span><strong>3 permissions</strong></div>
            <div className="accessItem"><span>runtime-service</span><strong>read scoped</strong></div>
            <div className="accessItem"><span>auditor</span><strong>no values</strong></div>
          </div>

          <div id="audit" className="panel compact">
            <h2>Audit stream</h2>
            {audit.map((event) => (
              <div className="auditItem" key={`${event.action}-${event.time}`}>
                <span className={event.outcome === "denied" ? "dot denied" : "dot"} />
                <div><strong>{event.action}</strong><small>{event.actor} to {event.target}</small></div>
                <time>{event.time}</time>
              </div>
            ))}
          </div>
        </section>

        <section id="cli" className="panel terminalPanel">
          <div>
            <h2>Runtime injection</h2>
            <p>Use short-lived API access from local tools without writing secret values to project files.</p>
          </div>
          <pre>devsvault run --path acme/api/prod/DATABASE_URL --name DATABASE_URL -- go run ./cmd/api</pre>
        </section>
      </section>
    </main>
  );
}