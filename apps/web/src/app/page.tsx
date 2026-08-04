import Image from "next/image";
import Link from "next/link";

const productSteps = [
  {
    title: "Send proposed actions",
    body: "n8n, LangGraph, custom agents, and backend services submit the action they want to perform before touching the business system.",
  },
  {
    title: "Evaluate clear policy",
    body: "HumanGate checks deterministic rules and returns allow, block, or approval required. No model confidence becomes an authorization decision.",
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
    <main className="min-h-screen bg-[#f6f0e6] text-[#13110d]">
      <section className="relative min-h-[88vh] overflow-hidden bg-[#101916] text-white">
        <Image
          alt="HumanGate approval inbox showing AI workflow actions awaiting review"
          className="absolute inset-0 h-full w-full object-cover object-center"
          height={992}
          priority
          src="/images/humangate-approval-inbox.png"
          width={1586}
        />
        <div className="absolute inset-0 bg-[linear-gradient(90deg,rgba(13,18,16,0.88)_0%,rgba(13,18,16,0.68)_34%,rgba(13,18,16,0.18)_72%,rgba(13,18,16,0.10)_100%)]" />
        <div className="absolute inset-x-0 bottom-0 h-24 bg-[linear-gradient(180deg,rgba(246,240,230,0)_0%,#f6f0e6_100%)]" />

        <div className="relative z-10 mx-auto flex min-h-[88vh] max-w-7xl flex-col px-5 py-5 sm:px-8 lg:px-10">
          <header className="flex items-center justify-between gap-4">
            <Link className="text-lg font-semibold" href="/">
              HumanGate
            </Link>
            <nav
              aria-label="Primary navigation"
              className="flex items-center gap-2 text-sm font-medium"
            >
              <Link
                className="hidden rounded-md px-3 py-2 text-white/80 transition hover:text-white sm:inline-flex"
                href="/login"
              >
                Sign in
              </Link>
              <Link
                className="rounded-md bg-white px-4 py-2 text-[#101916] transition hover:bg-[#e7f5f2]"
                href="/register"
              >
                Start
              </Link>
            </nav>
          </header>

          <div className="flex flex-1 items-center pb-20 pt-16">
            <div className="max-w-3xl">
              <p className="text-sm font-semibold uppercase tracking-[0.18em] text-[#8fd3ca]">
                Approval control for AI automation
              </p>
              <h1 className="mt-5 max-w-2xl text-5xl font-semibold leading-[1.02] sm:text-6xl lg:text-7xl">
                Human approval for AI workflows
              </h1>
              <p className="mt-6 max-w-xl text-lg leading-8 text-white/[0.82]">
                Put a policy gate between autonomous workflows and sensitive
                business actions. Let safe actions continue, block dangerous
                ones, and route risky decisions to the right person.
              </p>
              <div className="mt-8 flex flex-col gap-3 sm:flex-row">
                <Link
                  className="inline-flex min-h-12 items-center justify-center rounded-md bg-[#b74b2a] px-5 text-sm font-semibold text-white transition hover:bg-[#983b1f]"
                  href="/register"
                >
                  Create workspace
                </Link>
                <Link
                  className="inline-flex min-h-12 items-center justify-center rounded-md border border-white/40 px-5 text-sm font-semibold text-white transition hover:border-white hover:bg-white/10"
                  href="/login"
                >
                  Open dashboard
                </Link>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="relative z-10 mx-auto -mt-8 max-w-7xl px-5 pb-16 sm:px-8 lg:px-10">
        <div className="grid gap-4 md:grid-cols-3">
          {audiencePoints.map((point) => (
            <article
              className="rounded-lg border border-black/10 bg-white p-5 shadow-[0_18px_45px_rgba(0,0,0,0.08)]"
              key={point}
            >
              <p className="text-sm leading-6 text-black/70">{point}</p>
            </article>
          ))}
        </div>
      </section>

      <section className="border-y border-black/10 bg-[#fffaf1]">
        <div className="mx-auto grid max-w-7xl gap-10 px-5 py-16 sm:px-8 lg:grid-cols-[0.85fr_1.15fr] lg:px-10">
          <div>
            <p className="text-sm font-semibold uppercase tracking-[0.18em] text-[#1f6f78]">
              What it does
            </p>
            <h2 className="mt-4 text-3xl font-semibold sm:text-4xl">
              A control layer for the moment before an agent acts.
            </h2>
            <p className="mt-5 text-base leading-7 text-black/65">
              HumanGate does not issue refunds, delete data, publish content,
              or update your CRM. It reviews the proposed action, records the
              decision, and sends that decision back to the workflow that asked.
            </p>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            {productSteps.map((step, index) => (
              <article
                className="rounded-lg border border-black/10 bg-white p-5"
                key={step.title}
              >
                <p className="text-sm font-semibold text-[#b74b2a]">
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
          <p className="text-sm font-semibold uppercase tracking-[0.18em] text-[#1f6f78]">
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
                className="rounded-md border border-[#1f6f78]/20 bg-[#e9f7f5] px-3 py-2 text-sm font-semibold text-[#13545c]"
                key={integration}
              >
                {integration}
              </span>
            ))}
          </div>
        </div>

        <div className="rounded-lg border border-black/10 bg-[#152923] p-6 text-white">
          <p className="text-sm font-semibold uppercase tracking-[0.18em] text-[#8fd3ca]">
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
            <div className="rounded-md bg-[#8fd3ca] p-4 font-semibold text-[#10201c]">
              Workflow receives signed decision and resumes
            </div>
          </div>
        </div>
      </section>

      <section className="bg-[#12110e] px-5 py-14 text-white sm:px-8 lg:px-10">
        <div className="mx-auto flex max-w-7xl flex-col gap-6 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="text-3xl font-semibold">
              Add human judgement where automation needs it.
            </h2>
            <p className="mt-3 max-w-2xl text-sm leading-6 text-white/[0.68]">
              Start with a workspace, create an API key, and route your first
              proposed action through HumanGate.
            </p>
          </div>
          <Link
            className="inline-flex min-h-12 shrink-0 items-center justify-center rounded-md bg-white px-5 text-sm font-semibold text-[#12110e] transition hover:bg-[#e7f5f2]"
            href="/register"
          >
            Get started
          </Link>
        </div>
      </section>
    </main>
  );
}
