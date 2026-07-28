"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { getSupabaseBrowserClient } from "@/lib/supabase/client";

export function LoginForm() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);
    const supabase = getSupabaseBrowserClient();

    const { error: authError } = await supabase.auth.signInWithPassword({
      email,
      password,
    });

    setIsSubmitting(false);

    if (authError) {
      setError(authError.message);
      return;
    }

    router.push("/dashboard");
    router.refresh();
  }

  return (
    <form className="space-y-4" onSubmit={handleSubmit}>
      <label className="block space-y-2">
        <span className="text-sm font-medium text-[#15110d]">Email</span>
        <input
          required
          autoComplete="email"
          className="w-full rounded-2xl border border-black/10 bg-[#fbf8f2] px-4 py-3 outline-none transition focus:border-[#1f6f78] focus:ring-2 focus:ring-[#1f6f78]/15"
          onChange={(event) => setEmail(event.target.value)}
          type="email"
          value={email}
        />
      </label>

      <label className="block space-y-2">
        <span className="text-sm font-medium text-[#15110d]">Password</span>
        <input
          required
          autoComplete="current-password"
          className="w-full rounded-2xl border border-black/10 bg-[#fbf8f2] px-4 py-3 outline-none transition focus:border-[#1f6f78] focus:ring-2 focus:ring-[#1f6f78]/15"
          onChange={(event) => setPassword(event.target.value)}
          type="password"
          value={password}
        />
      </label>

      {error ? (
        <p className="rounded-2xl border border-[#b74b2a]/20 bg-[#fff2ec] px-4 py-3 text-sm text-[#8d3419]">
          {error}
        </p>
      ) : null}

      <button
        className="w-full rounded-2xl bg-[#15110d] px-4 py-3 text-sm font-semibold text-white transition hover:bg-black disabled:cursor-not-allowed disabled:opacity-60"
        disabled={isSubmitting}
        type="submit"
      >
        {isSubmitting ? "Signing in..." : "Sign in"}
      </button>
    </form>
  );
}
