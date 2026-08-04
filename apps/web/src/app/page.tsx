import Image from "next/image";
import Link from "next/link";

const productSteps = [
  {
    title: "Send proposed actions",
    body: "n8n, LangGraph, custom agents, and backend services submit the action they want to perform before touching the business system.",
  },
  {
    title: "Evaluate clear policy",
    body: "Greenpost checks deterministic rules and returns allow, block, or approval required. No model confidence becomes an authorization decision.",
  },
  {
    title: "Resolve in one inbox",
    body: "Approvers review the action, evidence, risk, source workflow, and deadline, then approve, edit and approve, or reject.",
  },
  {
    title: "Return a signed decision",
    body: "The original workflow receives a signed decision and remains responsible for the final action, retry behavior, and business credentials.",
  },
];

const audiencePoints = [
  "AI workflow builders who need a human checkpoint before refunds, deletions, customer messages, supplier changes, or publishing.",
  "Operations teams that want one approval inbox instead of scattered Slack threads, email chains, and one-off webhook forms.",
  "Engineering teams that need idempotent requests, signed callbacks, audit events, and workspace-scoped authorization.",
];

const integrations = ["n8n", "LangGraph", "Go services", "Python agents", "Generic webhooks"];

export default function HomePage() {
  return (
    <main className="min-h-screen bg-[#f4f7f1] text-[#111713]">
      <section className="relative min-h-[88vh] overflow-hidden bg-[#09110d] text-white">
        <Image
          alt="Greenpost approval inbox showing AI workflow actions awaiting review"
          className="absolute inset-0 h-full w-full object-cover object-[57%_center]"
          height={992}
          priority
          src="/images/greenpost-approval-inbox.png"
          width={1586}
        />
        <div className="absolute inset-0 bg-[linear-gradient(90deg,rgba(6,17,12,0.94)_0%,rgba(6,17,12,0.82)_40%,rgba(6,17,12,0.28)_76%,rgba(6,17,12,0.12)_100%)]" />
        <div className="absolute inset-x-0 bottom-0 h-24 bg-[linear-gradient(180deg,rgba(244,247,241,0)_0%,#f4f7f1_100%)]" />

        <div className="relative z-10 mx-auto flex min-h-[88vh] max-w-7xl flex-col px-5 py-5 sm:px-8 lg:px-10">
          <header className="flex items-center justify-between gap-4 rounded-lg border border-white/[0.12] bg-[#06110c]/75 px-3 py-3 shadow-[0_18px_55px_rgba(0,0,0,0.24)] backdrop-blur-md sm:px-4">
            <Link className="text-xl font-semibold" href="/">
              Greenpost
            </Link>
            <nav
              aria-label="Primary navigation"
              className="flex items-center gap-2 text-sm font-semibold"
            >
              <Link
                className="inline-flex min-h-11 items-center rounded-md border border-white/20 px-4 text-white transition hover:border-white/[0.45] hover:bg-white/10"
                href="/login"
              >
                Sign in
              </Link>
              <Link
                className="inline-flex min-h-11 items-center rounded-md bg-[#ccff66] px-4 text-[#0d170f] shadow-[0_10px_28px_rgba(204,255,102,0.22)] transition hover:bg-[#dcff88]"
                href="/register"
              >
                Get started
              </Link>
            </nav>
          </header>

          <div className="flex flex-1 items-center pb-20 pt-16">
            <div className="max-w-3xl">
              <p className="inline-flex rounded-md border border-[#ccff66]/30 bg-[#ccff66]/10 px-3 py-2 text-sm font-semibold text-[#dcff88]">
                For people building AI workflows
              </p>
              <h1 className="mt-5 max-w-2xl text-5xl font-semibold leading-[1.02] sm:text-6xl lg:text-7xl">
                Stop risky AI actions before they ship.
              </h1>
              <p className="mt-6 max-w-2xl text-lg leading-8 text-white/[0.84]">
                Greenpost is the approval inbox for n8n, LangGraph, and custom
                agents. Safe actions continue, blocked actions stop, and risky
                actions wait for a human decision.
              </p>
              <div className="mt-8 flex flex-col gap-3 sm:flex-row">
                <Link
                  className="inline-flex min-h-14 items-center justify-center rounded-md bg-[#ccff66] px-6 text-base font-semibold text-[#0d170f] shadow-[0_18px_45px_rgba(204,255,102,0.22)] transition hover:bg-[#dcff88]"
                  href="/register"
                >
                  Get started
                </Link>
                <Link
                  className="inline-flex min-h-14 items-center justify-center rounded-md border border-white/[0.35] bg-white/10 px-6 text-base font-semibold text-white backdrop-blur-sm transition hover:border-white/70 hover:bg-white/[0.16]"
                  href="/login"
                >
                  Sign in
                </Link>
              </div>
              <p className="mt-5 text-sm leading-6 text-white/[0.62]">
                No execution engine. No hidden business action. Just policy,
                approval, audit, and signed continuation.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="relative z-10 mx-auto -mt-8 max-w-7xl px-5 pb-16 sm:px-8 lg:px-10">
        <div className="grid gap-4 md:grid-cols-3">
          {audiencePoints.map((point) => (
            <article
              className="rounded-lg border border-[#152316]/10 bg-white p-5 shadow-[0_18px_45px_rgba(13,24,16,0.08)]"
              key={point}
            >
              <p className="text-sm leading-6 text-black/70">{point}</p>
            </article>
          ))}
        </div>
      </section>

      <section className="border-y border-[#152316]/10 bg-white">
        <div className="mx-auto grid max-w-7xl gap-10 px-5 py-16 sm:px-8 lg:grid-cols-[0.85fr_1.15fr] lg:px-10">
          <div>
            <p className="text-sm font-semibold uppercase tracking-[0.18em] text-[#2d6a4f]">
              What it does
            </p>
            <h2 className="mt-4 text-3xl font-semibold sm:text-4xl">
              A control layer for the moment before an agent acts.
            </h2>
            <p className="mt-5 text-base leading-7 text-black/65">
              Greenpost does not issue refunds, delete data, publish content,
              or update your CRM. It reviews the proposed action, records the
              decision, and sends that decision back to the workflow that asked.
            </p>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            {productSteps.map((step, index) => (
              <article
                className="rounded-lg border border-[#152316]/10 bg-[#f8fbf4] p-5"
                key={step.title}
              >
                <p className="text-sm font-semibold text-[#2d6a4f]">
                  0{index + 1}
                </p>
                <h3 className="mt-4 text-xl font-semibold">{step.title}</h3>
                <p className="mt-3 text-sm leading-6 text-black/65">
                  {step.body}
                </p>
              </article>
            ))}
          </div>
        </div>
      </section>

      <section className="mx-auto grid max-w-7xl gap-10 px-5 py-16 sm:px-8 lg:grid-cols-[1.05fr_0.95fr] lg:px-10">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.18em] text-[#2d6a4f]">
            Built for workflow creators
          </p>
          <h2 className="mt-4 text-3xl font-semibold sm:text-4xl">
            Keep automation moving without giving agents unchecked authority.
          </h2>
          <p className="mt-5 text-base leading-7 text-black/65">
            Use it anywhere an AI workflow can propose a structured action:
            agent tools, HTTP nodes, durable graph interruptions, internal
            services, and webhook-based automations.
          </p>
          <div className="mt-7 flex flex-wrap gap-2">
            {integrations.map((integration) => (
              <span
                className="rounded-md border border-[#2d6a4f]/20 bg-white px-3 py-2 text-sm font-semibold text-[#1e4f3b]"
                key={integration}
              >
                {integration}
              </span>
            ))}
          </div>
        </div>

        <div className="rounded-lg border border-[#152316]/10 bg-[#102017] p-6 text-white shadow-[0_18px_45px_rgba(13,24,16,0.12)]">
          <p className="text-sm font-semibold uppercase tracking-[0.18em] text-[#ccff66]">
            Typical gate
          </p>
          <div className="mt-6 space-y-3 text-sm">
            <div className="rounded-md bg-white/[0.08] p-4">
              Agent proposes: refund customer $249
            </div>
            <div className="rounded-md bg-white/[0.08] p-4">
              Policy matches: refunds over $50 require approval
            </div>
            <div className="rounded-md bg-white/[0.08] p-4">
              Approver reviews evidence, edits if needed, and signs off
            </div>
            <div className="rounded-md bg-[#ccff66] p-4 font-semibold text-[#102017]">
              Workflow receives signed decision and resumes
            </div>
          </div>
        </div>
      </section>

      <section className="bg-[#07110b] px-5 py-14 text-white sm:px-8 lg:px-10">
        <div className="mx-auto flex max-w-7xl flex-col gap-6 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="text-3xl font-semibold">
              Add human judgement where automation needs it.
            </h2>
            <p className="mt-3 max-w-2xl text-sm leading-6 text-white/[0.68]">
              Start with a workspace, create an API key, and route your first
              proposed action through Greenpost.
            </p>
          </div>
          <Link
            className="inline-flex min-h-12 shrink-0 items-center justify-center rounded-md bg-[#ccff66] px-5 text-sm font-semibold text-[#0d170f] transition hover:bg-[#dcff88]"
            href="/register"
          >
            Get started
          </Link>
        </div>
      </section>
    </main>
  );
}
