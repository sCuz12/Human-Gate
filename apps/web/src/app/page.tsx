import Image from "next/image";
import Link from "next/link";
import { IBM_Plex_Mono, IBM_Plex_Sans, Space_Grotesk } from "next/font/google";
import {
  ArrowRight,
  BracketsCurly,
  CheckCircle,
  Code,
  GitBranch,
  LockKey,
  PaperPlaneTilt,
  Plugs,
  ShieldCheck,
  Tray,
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

const proofLogos = [
  { label: "n8n", logo: "https://cdn.simpleicons.org/n8n/123d2b" },
  { label: "Go", logo: "https://cdn.simpleicons.org/go/123d2b" },
  { label: "Python", logo: "https://cdn.simpleicons.org/python/123d2b" },
  { label: "Slack", logo: "https://cdn.simpleicons.org/slack/123d2b" },
  { label: "Supabase", logo: "https://cdn.simpleicons.org/supabase/123d2b" },
];

const reviewCards = [
  {
    title: "Refund over policy limit",
    meta: "customer.refund",
    detail: "$249 proposed by an agent after a support conversation.",
    verdict: "Human review",
  },
  {
    title: "Supplier bank change",
    meta: "vendor.update",
    detail: "Irreversible account detail change before payment run.",
    verdict: "Blocked",
  },
  {
    title: "Knowledge base edit",
    meta: "content.publish",
    detail: "Low-risk article update with matching policy snapshot.",
    verdict: "Allowed",
  },
];

const bentoItems = [
  {
    title: "Policies decide first",
    body: "Deterministic rules return allow, block, or approval required before the workflow touches a business system.",
    icon: ShieldCheck,
    className: "md:col-span-2 md:row-span-2",
  },
  {
    title: "One inbox",
    body: "Approvers see action details, evidence, source workflow, deadline, and history in one place.",
    icon: Tray,
    className: "",
  },
  {
    title: "Signed callbacks",
    body: "Every resolved decision carries stable identifiers and hashes for replay-resistant continuation.",
    icon: LockKey,
    className: "",
  },
  {
    title: "Agent-native",
    body: "Submit from n8n, LangGraph interrupts, backend services, Python agents, or a generic webhook.",
    icon: GitBranch,
    className: "md:col-span-2",
  },
];

const flowSteps = [
  { title: "Propose", body: "Send the action and continuation target.", icon: PaperPlaneTilt },
  { title: "Evaluate", body: "Match active policy versions by priority.", icon: BracketsCurly },
  { title: "Review", body: "Approve, edit and approve, reject, or expire.", icon: CheckCircle },
  { title: "Resume", body: "Return a signed decision to the source workflow.", icon: ArrowRight },
];

export default function HomePage() {
  return (
    <main
      className={`${display.variable} ${body.variable} ${mono.variable} min-h-[100dvh] overflow-hidden bg-[#f6f8f3] font-[family-name:var(--font-body)] text-[#123d2b]`}
    >
      <section className="relative min-h-[100dvh] overflow-hidden bg-[#f6f8f3]">
        <Image
          alt="Abstract glass approval architecture with green illuminated decision panels"
          className="absolute inset-y-0 right-0 h-full w-full object-cover opacity-28 mix-blend-multiply lg:left-[34%] lg:w-[82%]"
          fill
          priority
          sizes="100vw"
          src="/images/decree-glass-gate.png"
        />
        <div className="absolute inset-0 bg-[linear-gradient(90deg,rgba(246,248,243,0.98)_0%,rgba(246,248,243,0.9)_43%,rgba(246,248,243,0.54)_100%)]" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_16%_76%,rgba(52,211,133,0.18),transparent_32%),radial-gradient(circle_at_86%_12%,rgba(195,213,201,0.45),transparent_28%)]" />
        <div className="gp-noise" />

        <div className="relative mx-auto flex min-h-[100dvh] max-w-7xl flex-col px-5 py-5 sm:px-8 lg:px-10">
          <header className="gp-hero-in flex h-16 items-center justify-between gap-4 rounded-2xl border border-[#123d2b]/10 bg-white/58 px-4 shadow-[0_18px_70px_rgba(18,61,43,0.1)] backdrop-blur-xl">
            <Link className="flex items-center gap-3" href="/">
              <span className="flex h-9 w-9 items-center justify-center rounded-xl border border-[#1f8f5f]/20 bg-[#dff7e8] text-[#123d2b]">
                <ShieldCheck size={19} weight="bold" />
              </span>
              <span className="font-[family-name:var(--font-display)] text-xl font-semibold tracking-tight">
                Decree
              </span>
            </Link>
            <nav aria-label="Primary navigation" className="flex items-center gap-2 text-sm font-semibold">
              <Link
                className="inline-flex min-h-11 items-center rounded-xl px-4 text-[#123d2b]/68 transition hover:bg-[#123d2b]/8 hover:text-[#123d2b] active:translate-y-px"
                href="/login"
              >
                Sign in
              </Link>
              <Link
                className="inline-flex min-h-11 items-center rounded-xl bg-[#123d2b] px-4 text-white shadow-[0_12px_35px_rgba(18,61,43,0.18)] transition hover:bg-[#19513a] active:translate-y-px"
                href="/register"
              >
                Get started
              </Link>
            </nav>
          </header>

          <div className="grid flex-1 items-center gap-10 py-12 lg:grid-cols-[minmax(0,0.92fr)_minmax(360px,0.58fr)] lg:py-10">
            <div className="max-w-4xl">
              <p className="gp-hero-in max-w-fit rounded-full border border-[#1f8f5f]/20 bg-white/62 px-4 py-2 font-[family-name:var(--font-mono)] text-xs font-medium text-[#1b7652] shadow-[0_10px_30px_rgba(18,61,43,0.08)] backdrop-blur-xl">
                Human approval for agentic workflows
              </p>
              <h1
                className="gp-hero-in mt-7 max-w-4xl font-[family-name:var(--font-display)] text-5xl font-semibold leading-[0.96] tracking-tight text-[#123d2b] sm:text-6xl lg:text-7xl"
                style={{ animationDelay: "80ms" }}
              >
                Turn agent actions into reviewed decisions.
              </h1>
              <p
                className="gp-hero-in mt-7 max-w-xl text-lg leading-8 text-[#365846]"
                style={{ animationDelay: "160ms" }}
              >
                Decree routes proposed actions through policy, people, audit, and signed continuation.
              </p>
              <div
                className="gp-hero-in mt-9 flex flex-col gap-3 sm:flex-row"
                style={{ animationDelay: "240ms" }}
              >
                <Link
                  className="inline-flex min-h-14 items-center justify-center gap-2 rounded-xl bg-[#123d2b] px-6 text-base font-semibold text-white shadow-[0_18px_45px_rgba(18,61,43,0.18)] transition hover:bg-[#19513a] active:translate-y-px"
                  href="/register"
                >
                  Get started
                  <ArrowRight size={18} weight="bold" />
                </Link>
                <Link
                  className="inline-flex min-h-14 items-center justify-center rounded-xl border border-[#123d2b]/15 bg-white/62 px-6 text-base font-semibold text-[#123d2b] shadow-[0_14px_35px_rgba(18,61,43,0.08)] backdrop-blur-md transition hover:border-[#123d2b]/32 hover:bg-white active:translate-y-px"
                  href="/login"
                >
                  Sign in
                </Link>
              </div>
            </div>

            <div className="gp-card-in hidden rounded-[2rem] border border-white/70 bg-white/64 p-3 shadow-[0_30px_100px_rgba(18,61,43,0.16)] backdrop-blur-xl lg:block">
              <div className="rounded-[1.5rem] border border-[#123d2b]/10 bg-[#fafffb]/82 p-4">
                <div className="flex items-center justify-between border-b border-[#123d2b]/10 pb-4">
                  <div>
                    <p className="font-[family-name:var(--font-display)] text-lg font-semibold">Live review queue</p>
                    <p className="mt-1 text-sm text-[#365846]/70">Policy has already sorted what needs a person.</p>
                  </div>
                  <span className="rounded-full bg-[#dff7e8] px-3 py-1 text-sm font-semibold text-[#1b7652]">
                    3 pending
                  </span>
                </div>
                <div className="mt-4 space-y-3">
                  {reviewCards.map((card, index) => (
                    <article
                      className="gp-float-panel rounded-2xl border border-[#123d2b]/10 bg-white/80 p-4 shadow-[0_18px_50px_rgba(18,61,43,0.08)]"
                      key={card.title}
                      style={{ animationDelay: `${index * 180}ms` }}
                    >
                      <div className="flex items-start justify-between gap-4">
                        <div>
                          <p className="font-[family-name:var(--font-mono)] text-xs text-[#1b7652]">{card.meta}</p>
                          <h2 className="mt-2 font-[family-name:var(--font-display)] text-lg font-semibold">
                            {card.title}
                          </h2>
                        </div>
                        <span className="shrink-0 rounded-full border border-[#1f8f5f]/18 bg-[#f2fff7] px-3 py-1 text-xs font-semibold text-[#1b7652]">
                          {card.verdict}
                        </span>
                      </div>
                      <p className="mt-3 text-sm leading-6 text-[#365846]/74">{card.detail}</p>
                    </article>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="border-y border-[#123d2b]/10 bg-white/54">
        <div className="mx-auto flex max-w-7xl flex-wrap items-center justify-center gap-x-10 gap-y-6 px-5 py-7 sm:px-8 lg:px-10">
          {proofLogos.map((item) => (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              alt={item.label}
              className="h-7 w-auto opacity-55 grayscale transition hover:opacity-100 hover:grayscale-0"
              height={28}
              key={item.label}
              src={item.logo}
              width={112}
            />
          ))}
        </div>
      </section>

      <section className="relative bg-[#f6f8f3] px-5 py-20 sm:px-8 lg:px-10">
        <div className="mx-auto max-w-7xl">
          <ScrollReveal className="max-w-3xl">
            <h2 className="font-[family-name:var(--font-display)] text-4xl font-semibold leading-tight tracking-tight text-[#123d2b] sm:text-5xl">
              Built for the moment before a workflow acts.
            </h2>
            <p className="mt-5 max-w-2xl text-base leading-7 text-[#365846]">
              The platform never performs the final business action. It decides whether the source workflow may continue.
            </p>
          </ScrollReveal>

          <div className="mt-12 grid auto-rows-[minmax(220px,auto)] gap-4 md:grid-cols-4">
            {bentoItems.map((item, index) => (
              <ScrollReveal className={item.className} delayMs={index * 90} key={item.title}>
                <article className="group relative flex h-full min-h-[220px] overflow-hidden rounded-[1.5rem] border border-[#123d2b]/10 bg-white p-6 shadow-[0_24px_80px_rgba(18,61,43,0.08)]">
                  <div className="absolute inset-0 opacity-0 transition duration-500 group-hover:opacity-100 [background:radial-gradient(circle_at_70%_20%,rgba(52,211,133,0.16),transparent_34%)]" />
                  <div className="relative flex h-full flex-col justify-between">
                    <span className="flex h-11 w-11 items-center justify-center rounded-2xl border border-[#1f8f5f]/16 bg-[#dff7e8] text-[#123d2b]">
                      <item.icon size={22} weight="bold" />
                    </span>
                    <div className="mt-10">
                      <h3 className="font-[family-name:var(--font-display)] text-2xl font-semibold tracking-tight">
                        {item.title}
                      </h3>
                      <p className="mt-3 max-w-lg text-sm leading-6 text-[#365846]/76">{item.body}</p>
                    </div>
                  </div>
                </article>
              </ScrollReveal>
            ))}
            <ScrollReveal className="md:col-span-2" delayMs={360}>
              <article className="relative h-full min-h-[220px] overflow-hidden rounded-[1.5rem] border border-[#1f8f5f]/18 bg-[#dff7e8] p-6">
                <div className="absolute inset-0 [background:linear-gradient(135deg,rgba(255,255,255,0.72),transparent_48%),radial-gradient(circle_at_85%_70%,rgba(52,211,133,0.24),transparent_30%)]" />
                <div className="relative max-w-xl">
                  <p className="font-[family-name:var(--font-mono)] text-sm text-[#1b7652]">Decision integrity</p>
                  <h3 className="mt-6 font-[family-name:var(--font-display)] text-3xl font-semibold leading-tight">
                    Original action hash. Approved action hash. One immutable audit trail.
                  </h3>
                </div>
              </article>
            </ScrollReveal>
          </div>
        </div>
      </section>

      <section className="relative overflow-hidden bg-[#edf3ed] px-5 py-20 sm:px-8 lg:px-10">
        <div className="mx-auto grid max-w-7xl gap-10 lg:grid-cols-[0.9fr_1.1fr] lg:items-center">
          <ScrollReveal>
            <h2 className="font-[family-name:var(--font-display)] text-4xl font-semibold leading-tight tracking-tight sm:text-5xl">
              The inbox becomes the control surface.
            </h2>
            <p className="mt-5 max-w-xl text-base leading-7 text-[#365846]">
              Review proposed action, policy match, evidence, affected systems, and deadline before anything resumes.
            </p>
            <div className="mt-8 grid gap-3 sm:grid-cols-2">
              {[
                ["Workspace-scoped authorization", ShieldCheck],
                ["Edit and approve flows", Code],
                ["n8n and webhook continuations", Plugs],
                ["LangGraph-friendly idempotency", GitBranch],
              ].map(([label, Icon]) => (
                <div className="rounded-2xl border border-[#123d2b]/10 bg-white/70 p-4 text-sm font-semibold shadow-[0_18px_50px_rgba(18,61,43,0.06)]" key={label as string}>
                  <Icon className="mb-3 text-[#1b7652]" size={20} weight="bold" />
                  {label as string}
                </div>
              ))}
            </div>
          </ScrollReveal>

          <ScrollReveal delayMs={120}>
            <div className="relative overflow-hidden rounded-[2rem] border border-[#123d2b]/10 bg-white p-2 shadow-[0_36px_110px_rgba(18,61,43,0.14)]">
              <Image
                alt="Decree approval inbox interface for reviewing pending workflow actions"
                className="h-auto w-full rounded-[1.55rem]"
                height={992}
                sizes="(min-width: 1024px) 640px, 92vw"
                src="/images/decree-approval-inbox.png"
                width={1586}
              />
            </div>
          </ScrollReveal>
        </div>
      </section>

      <section className="bg-[#f6f8f3] px-5 py-20 sm:px-8 lg:px-10">
        <div className="mx-auto max-w-7xl">
          <ScrollReveal className="max-w-2xl">
            <h2 className="font-[family-name:var(--font-display)] text-4xl font-semibold tracking-tight sm:text-5xl">
              Four moves. No magic.
            </h2>
          </ScrollReveal>
          <div className="mt-12 grid gap-4 md:grid-cols-4">
            {flowSteps.map((step, index) => (
              <ScrollReveal delayMs={index * 80} key={step.title}>
                <article className="relative min-h-[245px] rounded-[1.5rem] border border-[#123d2b]/10 bg-white p-5 shadow-[0_24px_80px_rgba(18,61,43,0.08)]">
                  <step.icon className="text-[#1b7652]" size={24} weight="bold" />
                  <h3 className="mt-12 font-[family-name:var(--font-display)] text-2xl font-semibold">
                    {step.title}
                  </h3>
                  <p className="mt-3 text-sm leading-6 text-[#365846]/76">{step.body}</p>
                </article>
              </ScrollReveal>
            ))}
          </div>
        </div>
      </section>

      <section className="relative overflow-hidden bg-[#edf3ed] px-5 py-20 sm:px-8 lg:px-10">
        <div className="absolute inset-0 [background:radial-gradient(circle_at_50%_0%,rgba(52,211,133,0.2),transparent_42%)]" />
        <ScrollReveal className="relative mx-auto flex max-w-7xl flex-col items-start gap-8 rounded-[2rem] border border-white/70 bg-white/72 p-8 shadow-[0_30px_100px_rgba(18,61,43,0.12)] backdrop-blur-xl md:flex-row md:items-center md:justify-between md:p-10">
          <div>
            <h2 className="max-w-3xl font-[family-name:var(--font-display)] text-4xl font-semibold leading-tight tracking-tight sm:text-5xl">
              Put policy and people inside every workflow.
            </h2>
            <p className="mt-4 max-w-2xl text-base leading-7 text-[#365846]">
              Create a workspace, generate an API key, and route your first proposed action through Decree.
            </p>
          </div>
          <Link
            className="inline-flex min-h-14 shrink-0 items-center justify-center gap-2 rounded-xl bg-[#123d2b] px-6 text-base font-semibold text-white transition hover:bg-[#19513a] active:translate-y-px"
            href="/register"
          >
            Get started
            <ArrowRight size={18} weight="bold" />
          </Link>
        </ScrollReveal>
      </section>

      <footer className="border-t border-[#123d2b]/10 bg-[#f6f8f3] px-5 py-8 sm:px-8 lg:px-10">
        <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-4 text-sm text-[#365846]/70 sm:flex-row">
          <span>© {new Date().getFullYear()} Decree</span>
          <div className="flex items-center gap-5 font-medium">
            <Link className="transition hover:text-[#123d2b]" href="/login">
              Sign in
            </Link>
            <Link className="transition hover:text-[#123d2b]" href="/register">
              Get started
            </Link>
          </div>
        </div>
      </footer>
    </main>
  );
}
