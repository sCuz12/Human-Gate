import type { ReactNode } from "react";
import Link from "next/link";

type AuthCardProps = {
  eyebrow: string;
  title: string;
  description: string;
  alternateLabel: string;
  alternateHref: string;
  alternateText: string;
  children: ReactNode;
};

export function AuthCard({
  eyebrow,
  title,
  description,
  alternateLabel,
  alternateHref,
  alternateText,
  children,
}: AuthCardProps) {
  return (
    <div className="grid min-h-screen lg:grid-cols-[1.15fr_0.85fr]">
      <section className="hidden bg-[#16302b] px-10 py-12 text-[#f4ead8] lg:flex lg:flex-col lg:justify-between">
        <div className="space-y-6">
          <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#c8e0d9]">
            Greenpost
          </p>
          <h1 className="max-w-xl text-6xl font-semibold leading-[1.02]">
            Review consequential AI actions before they hit the business.
          </h1>
          <p className="max-w-lg text-lg leading-8 text-[#d4c9b8]">
            One workspace for policy checks, human approvals, and signed
            workflow continuation across n8n, LangGraph, and custom systems.
          </p>
        </div>

        <div className="grid gap-4">
          <div className="rounded-[1.75rem] border border-white/10 bg-white/5 p-5">
            <p className="text-sm uppercase tracking-[0.18em] text-[#9cc7bb]">
              What this unlocks
            </p>
            <p className="mt-3 text-base leading-7 text-[#efe6d9]">
              Central approval policies, a shared inbox, immutable audit
              history, and workflow-safe decisions.
            </p>
          </div>
        </div>
      </section>

      <section className="flex items-center justify-center bg-[#f7f1e7] px-6 py-12 md:px-10">
        <div className="w-full max-w-md rounded-[2rem] border border-black/10 bg-white px-6 py-8 shadow-[0_28px_80px_rgba(0,0,0,0.10)] md:px-8">
          <div className="space-y-3">
            <p className="text-sm font-semibold uppercase tracking-[0.22em] text-[#b74b2a]">
              {eyebrow}
            </p>
            <h2 className="text-3xl font-semibold text-[#15110d]">{title}</h2>
            <p className="text-sm leading-6 text-black/65">{description}</p>
          </div>

          <div className="mt-8">{children}</div>

          <p className="mt-6 text-sm text-black/60">
            {alternateLabel}{" "}
            <Link className="font-medium text-[#1f6f78]" href={alternateHref}>
              {alternateText}
            </Link>
          </p>
        </div>
      </section>
    </div>
  );
}
