import { Fragment } from "react";
import Image from "next/image";
import Link from "next/link";
import { IBM_Plex_Mono, IBM_Plex_Sans, Space_Grotesk } from "next/font/google";
import {
  CaretRight,
  Code,
  FlowArrow,
  GitBranch,
  PaperPlaneTilt,
  Plugs,
  SealCheck,
  ShieldCheck,
  Tray,
  TrayArrowDown,
} from "@phosphor-icons/react/ssr";
import { ScrollReveal } from "@/components/landing/scroll-reveal";

const display = Space_Grotesk({
  subsets: ["latin"],
  weight: ["500", "600", "700"],
  variable: "--font-display",
  display: "swap",
});

const body = IBM_Plex_Sans({
  subsets: ["latin"],
  weight: ["400", "500", "600"],
  variable: "--font-body",
  display: "swap",
});

const mono = IBM_Plex_Mono({
  subsets: ["latin"],
  weight: ["400", "500"],
  variable: "--font-mono",
  display: "swap",
});

const audiencePoints = [
  {
    tag: "Workflow builders",
    body: "Add a human checkpoint before refunds, deletions, customer messages, supplier changes, or publishing.",
    icon: FlowArrow,
  },
  {
    tag: "Operations",
    body: "Work every approval from one inbox instead of scattered Slack threads, email chains, and one-off webhook forms.",
    icon: Tray,
  },
  {
    tag: "Engineering",
    body: "Get idempotent requests, signed callbacks, audit events, and approvals scoped to the right workspace.",
    icon: Code,
  },
];

const productSteps = [
  {
    title: "Send the proposed action",
    body: "n8n, LangGraph, custom agents, and backend services submit the action they want to take before it touches a business system.",
    icon: PaperPlaneTilt,
  },
  {
    title: "Check the policy",
    body: "Decree runs deterministic rules and returns allow, block, or approval required. Model confidence never becomes an authorization decision.",
    icon: ShieldCheck,
  },
  {
    title: "Resolve it in one inbox",
    body: "An approver reviews the action, evidence, risk, and deadline, then approves, edits and approves, or rejects it.",
    icon: TrayArrowDown,
  },
  {
    title: "Get a signed decision back",
    body: "The workflow receives a signed decision and stays responsible for the actual execution, retries, and credentials.",
    icon: SealCheck,
  },
];

const integrations = [
  { label: "n8n", logo: "https://cdn.simpleicons.org/n8n/13505a" },
  { label: "LangGraph", icon: GitBranch },
  { label: "Go services", logo: "https://cdn.simpleicons.org/go/13505a" },
  { label: "Python agents", logo: "https://cdn.simpleicons.org/python/13505a" },
  { label: "Generic webhooks", icon: Plugs },
];

const gateLines = [
  { text: "Agent proposes: refund customer $249", tone: "default" },
  { text: "Policy matches: refunds over $50 require approval", tone: "hold" },
  { text: "Approver reviews evidence, edits if needed, and signs off", tone: "default" },
  { text: "Workflow receives its signed decision and resumes", tone: "go" },
] as const;

export default function HomePage() {
  const LeadAudienceIcon = audiencePoints[0].icon;

  return (
    <main
      className={`${display.variable} ${body.variable} ${mono.variable} min-h-screen bg-[#f7f1e7] font-[family-name:var(--font-body)] text-[#15110d]`}
    >
      {/* ---------- HERO (asymmetric split, real screenshot as the visual) ---------- */}
      <section className="relative overflow-hidden bg-[#16302b] text-white">
        <div className="pointer-events-none absolute inset-0 [background:radial-gradient(60%_50%_at_85%_0%,rgba(23,106,68,0.32),transparent_60%)]" />

        <div className="relative z-10 mx-auto max-w-7xl px-5 py-5 sm:px-8 lg:px-10">
          <header className="gp-hero-in flex items-center justify-between gap-4 rounded-lg border border-white/[0.12] bg-[#16302b]/80 px-3 py-3 backdrop-blur-md sm:px-4">
            <Link className="flex items-center gap-2" href="/">
              <svg aria-hidden="true" className="h-6 w-6 text-white/80" fill="none" viewBox="0 0 24 24">
                <circle cx="12" cy="12" r="9" stroke="currentColor" strokeDasharray="2.4 3.2" strokeWidth="1.5" />
                <circle cx="12" cy="12" r="2.5" fill="#4fb583" />
              </svg>
              <span className="font-[family-name:var(--font-display)] text-xl font-semibold tracking-tight">
                Decree
              </span>
            </Link>
            <nav aria-label="Primary navigation" className="flex items-center gap-2 text-sm font-semibold">
              <Link
                className="inline-flex min-h-11 items-center rounded-md border border-white/20 px-4 text-white transition hover:border-white/45 hover:bg-white/10"
                href="/login"
              >
                Sign in
              </Link>
              <Link
                className="inline-flex min-h-11 items-center rounded-md bg-[#176a44] px-4 text-white transition hover:bg-[#0f5636]"
                href="/register"
              >
                Get started
              </Link>
            </nav>
          </header>

          <div className="grid gap-14 py-14 lg:grid-cols-[1fr_1fr] lg:items-center lg:py-20">
            <div className="max-w-lg">
              <p className="gp-hero-in inline-flex rounded-md border border-[#b74b2a]/40 bg-[#b74b2a]/15 px-3 py-2 text-sm font-semibold text-[#f3c3a8]">
                For people building AI workflows
              </p>
              <h1
                className="gp-hero-in mt-6 font-[family-name:var(--font-display)] text-4xl font-semibold leading-[1.1] sm:text-5xl lg:text-6xl"
                style={{ animationDelay: "80ms" }}
              >
                Give AI agents a checkpoint, not a blank check.
              </h1>
              <p
                className="gp-hero-in mt-6 text-lg leading-8 text-white/80"
                style={{ animationDelay: "160ms" }}
              >
                Decree is the approval post for n8n, LangGraph, and custom
                agents. Safe actions continue. Risky ones wait for a person.
              </p>
              <div
                className="gp-hero-in mt-9 flex flex-col gap-3 sm:flex-row"
                style={{ animationDelay: "240ms" }}
              >
                <Link
                  className="inline-flex min-h-14 items-center justify-center rounded-md bg-[#176a44] px-6 text-base font-semibold text-white transition hover:bg-[#0f5636] active:translate-y-px"
                  href="/register"
                >
                  Get started
                </Link>
                <Link
                  className="inline-flex min-h-14 items-center justify-center rounded-md border border-white/35 bg-white/10 px-6 text-base font-semibold text-white backdrop-blur-sm transition hover:border-white/70 hover:bg-white/[0.16] active:translate-y-px"
                  href="/login"
                >
                  Sign in
                </Link>
              </div>
            </div>

            {/* Signature element: a real screenshot, stamped cleared */}
            <div className="relative mx-auto w-full max-w-lg lg:mx-0">
              <div
                className="gp-manifest-in overflow-hidden rounded-xl border border-black/10 bg-white shadow-[0_30px_80px_rgba(0,0,0,0.4)]"
                style={{ animationDelay: "260ms" }}
              >
                <div className="flex items-center gap-1.5 border-b border-black/10 bg-[#f4f5f0] px-4 py-3">
                  <span className="h-2.5 w-2.5 rounded-full bg-[#b74b2a]/50" />
                  <span className="h-2.5 w-2.5 rounded-full bg-[#b68422]/50" />
                  <span className="h-2.5 w-2.5 rounded-full bg-[#176a44]/50" />
                </div>
                <Image
                  alt="Decree approval inbox showing an AI workflow action awaiting review"
                  className="h-auto w-full"
                  height={992}
                  priority
                  sizes="(min-width: 1024px) 520px, 90vw"
                  src="/images/decree-approval-inbox.png"
                  width={1586}
                />
              </div>

              <div
                className="gp-stamp-in pointer-events-none absolute left-[86%] top-[2%] flex h-24 w-24 items-center justify-center rounded-full border-[3px] border-[#176a44] bg-[#faf6ee] text-[#176a44] shadow-[0_10px_30px_rgba(0,0,0,0.25)]"
                style={{ animationDelay: "820ms" }}
              >
                <div className="flex flex-col items-center leading-none">
                  <span className="font-[family-name:var(--font-mono)] text-[8px] tracking-[0.2em]">
                    GREENPOST
                  </span>
                  <span className="mt-1 font-[family-name:var(--font-display)] text-sm font-semibold tracking-wide">
                    CLEARED
                  </span>
                  <span className="mt-1 font-[family-name:var(--font-mono)] text-[8px] tracking-[0.2em]">
                    SIGNED
                  </span>
                </div>
              </div>

              <div
                className="gp-hero-in mt-5 flex items-center gap-5 text-xs text-white/55"
                style={{ animationDelay: "1000ms" }}
              >
                <span className="flex items-center gap-1.5">
                  <span className="h-1.5 w-1.5 rounded-full bg-white/25" />
                  Block
                </span>
                <span className="flex items-center gap-1.5">
                  <span className="h-1.5 w-1.5 rounded-full bg-[#b68422]/70" />
                  Hold
                </span>
                <span className="flex items-center gap-1.5 text-[#7fd1a6]">
                  <span className="gp-lamp-go h-1.5 w-1.5 rounded-full bg-[#4fb583]" />
                  Go
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ---------- AUDIENCE (1 lead card + 2, not three equal columns) ---------- */}
      <section className="mx-auto max-w-7xl px-5 py-16 sm:px-8 lg:px-10">
        <div className="grid gap-4 md:grid-cols-2">
          <ScrollReveal className="md:col-span-2">
            <article className="rounded-lg border border-black/[0.08] bg-white p-6 shadow-[0_18px_45px_rgba(13,24,16,0.06)] transition duration-300 hover:-translate-y-1 hover:shadow-[0_24px_55px_rgba(13,24,16,0.1)] md:flex md:items-center md:gap-10">
              <div className="flex items-center gap-3 md:w-56 md:shrink-0">
                <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-[#1f6f78]/10 text-[#1f6f78]">
                  <LeadAudienceIcon size={18} weight="bold" />
                </span>
                <p className="text-base font-semibold text-[#1f6f78]">
                  {audiencePoints[0].tag}
                </p>
              </div>
              <p className="mt-3 text-sm leading-6 text-black/70 md:mt-0 md:max-w-lg">
                {audiencePoints[0].body}
              </p>
            </article>
          </ScrollReveal>
          {audiencePoints.slice(1).map((point, index) => (
            <ScrollReveal delayMs={(index + 1) * 100} key={point.tag}>
              <article className="h-full rounded-lg border border-black/[0.08] bg-white p-5 shadow-[0_18px_45px_rgba(13,24,16,0.06)] transition duration-300 hover:-translate-y-1 hover:shadow-[0_24px_55px_rgba(13,24,16,0.1)]">
                <span className="flex h-9 w-9 items-center justify-center rounded-md bg-[#1f6f78]/10 text-[#1f6f78]">
                  <point.icon size={18} weight="bold" />
                </span>
                <p className="mt-3 text-sm font-semibold text-[#1f6f78]">{point.tag}</p>
                <p className="mt-3 text-sm leading-6 text-black/70">{point.body}</p>
              </article>
            </ScrollReveal>
          ))}
        </div>
      </section>

      {/* ---------- WHAT IT DOES (stacked header, full-width process grid) ---------- */}
      <section className="bg-[#efe7dc]">
        <div className="mx-auto max-w-7xl px-5 py-16 sm:px-8 lg:px-10">
          <ScrollReveal className="max-w-2xl">
            <h2 className="font-[family-name:var(--font-display)] text-3xl font-semibold sm:text-4xl">
              A checkpoint for the moment before an agent acts.
            </h2>
            <p className="mt-5 text-base leading-7 text-black/65">
              Decree doesn&apos;t issue refunds, delete records, publish
              content, or touch your CRM. It reviews the proposed action,
              records the decision, and hands that decision back to the
              workflow that asked.
            </p>
          </ScrollReveal>

          <div className="mt-10 grid items-center gap-3 sm:grid-cols-2 lg:grid-cols-[1fr_auto_1fr_auto_1fr_auto_1fr]">
            {productSteps.map((step, index) => (
              <Fragment key={step.title}>
                <ScrollReveal delayMs={index * 90}>
                  <article className="h-full rounded-lg border border-black/[0.08] bg-white p-5 transition duration-300 hover:-translate-y-1 hover:shadow-[0_18px_40px_rgba(13,24,16,0.08)]">
                    <div className="flex items-center justify-between">
                      <p className="font-[family-name:var(--font-mono)] text-sm font-semibold text-[#176a44]">
                        0{index + 1}
                      </p>
                      <span className="flex h-8 w-8 items-center justify-center rounded-md bg-[#176a44]/10 text-[#176a44]">
                        <step.icon size={16} weight="bold" />
                      </span>
                    </div>
                    <h3 className="mt-4 text-lg font-semibold">{step.title}</h3>
                    <p className="mt-3 text-sm leading-6 text-black/65">{step.body}</p>
                  </article>
                </ScrollReveal>
                {index < productSteps.length - 1 ? (
                  <CaretRight
                    aria-hidden="true"
                    className="mx-auto hidden shrink-0 text-black/20 lg:block"
                    size={18}
                    weight="bold"
                  />
                ) : null}
              </Fragment>
            ))}
          </div>
        </div>
      </section>

      {/* ---------- WHERE IT PLUGS IN (full-width stacked, no split) ---------- */}
      <section className="mx-auto max-w-7xl px-5 py-16 sm:px-8 lg:px-10">
        <ScrollReveal className="max-w-2xl">
          <h2 className="font-[family-name:var(--font-display)] text-3xl font-semibold sm:text-4xl">
            Keep automations moving without handing agents a blank check.
          </h2>
          <p className="mt-5 text-base leading-7 text-black/65">
            Use Decree anywhere a workflow can propose a structured
            action: agent tools, HTTP nodes, durable graph interruptions,
            internal services, and webhooks.
          </p>
        </ScrollReveal>

        <ScrollReveal className="mt-7 flex flex-wrap gap-2" delayMs={80}>
          {integrations.map((integration) => (
            <span
              className="flex items-center gap-2 rounded-md border border-[#1f6f78]/20 bg-white px-3 py-2 text-sm font-semibold text-[#13505a] transition hover:-translate-y-0.5 hover:border-[#1f6f78]/40"
              key={integration.label}
            >
              {"logo" in integration ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img alt="" aria-hidden="true" className="h-4 w-4" height={16} src={integration.logo} width={16} />
              ) : (
                <integration.icon aria-hidden="true" size={16} weight="bold" />
              )}
              {integration.label}
            </span>
          ))}
        </ScrollReveal>

        <ScrollReveal className="mt-10" delayMs={140}>
          <div className="rounded-lg border border-black/10 bg-[#16302b] p-6 text-white">
            <div className="space-y-3 text-sm">
              {gateLines.map((line, index) => (
                <ScrollReveal delayMs={index * 130} key={line.text}>
                  <div
                    className={
                      line.tone === "go"
                        ? "gp-glow-pulse rounded-md bg-[#176a44] p-4 font-semibold text-white"
                        : line.tone === "hold"
                          ? "rounded-md border border-[#b74b2a]/30 bg-white/[0.06] p-4 text-white/90"
                          : "rounded-md bg-white/[0.08] p-4 text-white/90"
                    }
                  >
                    {line.text}
                  </div>
                </ScrollReveal>
              ))}
            </div>
          </div>
        </ScrollReveal>
      </section>

      {/* ---------- FINAL CTA ---------- */}
      <section className="relative overflow-hidden bg-[#16302b] px-5 py-16 text-white sm:px-8 lg:px-10">
        <div className="pointer-events-none absolute inset-0 [background:radial-gradient(50%_60%_at_10%_100%,rgba(23,106,68,0.28),transparent_60%)]" />
        <ScrollReveal className="relative z-10 mx-auto flex max-w-7xl flex-col gap-6 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="font-[family-name:var(--font-display)] text-3xl font-semibold sm:text-4xl">
              Put a checkpoint in front of your agents.
            </h2>
            <p className="mt-3 max-w-2xl text-sm leading-6 text-white/65">
              Create a workspace, generate an API key, and route your first
              proposed action through Decree.
            </p>
          </div>
          <Link
            className="inline-flex min-h-12 shrink-0 items-center justify-center rounded-md bg-[#176a44] px-5 text-sm font-semibold text-white transition hover:bg-[#0f5636] active:translate-y-px"
            href="/register"
          >
            Get started
          </Link>
        </ScrollReveal>
      </section>

      {/* ---------- FOOTER ---------- */}
      <footer className="border-t border-black/10 bg-[#f7f1e7] px-5 py-8 sm:px-8 lg:px-10">
        <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-3 text-sm sm:flex-row">
          <span className="text-black/45">© {new Date().getFullYear()} Decree</span>
          <div className="flex items-center gap-5 font-medium">
            <Link className="text-black/60 transition hover:text-[#1f6f78]" href="/login">
              Sign in
            </Link>
            <Link className="text-black/60 transition hover:text-[#1f6f78]" href="/register">
              Get started
            </Link>
          </div>
        </div>
      </footer>
    </main>
  );
}
