"use client";

import { Check, Copy } from "@phosphor-icons/react";
import { useState } from "react";

type CopySnippetProps = {
  label: string;
  value: string;
  language?: string;
};

export function CopySnippet({
  label,
  value,
  language = "text",
}: CopySnippetProps) {
  const [copied, setCopied] = useState(false);

  async function copyValue() {
    await navigator.clipboard.writeText(value);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1800);
  }

  return (
    <div className="overflow-hidden rounded-lg border border-black/10 bg-[#15110d] shadow-[0_16px_40px_rgba(0,0,0,0.08)]">
      <div className="flex items-center justify-between gap-3 border-b border-white/10 bg-white/5 px-4 py-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#9bc5bb]">
            {label}
          </p>
          <p className="mt-1 text-xs text-[#dbcdbb]">{language}</p>
        </div>
        <button
          className="inline-flex items-center gap-2 rounded-md border border-white/15 bg-white px-3 py-2 text-sm font-semibold text-[#15110d] transition hover:bg-[#f5ecde] active:translate-y-px"
          onClick={copyValue}
          type="button"
        >
          {copied ? <Check aria-hidden size={16} weight="bold" /> : <Copy aria-hidden size={16} />}
          {copied ? "Copied" : "Copy"}
        </button>
      </div>
      <pre className="max-h-[520px] overflow-x-auto p-4 text-sm leading-6 text-[#f5ecde]">
        <code>{value}</code>
      </pre>
    </div>
  );
}
