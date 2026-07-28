"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { getSupabaseBrowserClient } from "@/lib/supabase/client";

type AuthRedirectProps = {
  toAuthenticated: string;
  toUnauthenticated: string;
};

export function AuthRedirect({
  toAuthenticated,
  toUnauthenticated,
}: AuthRedirectProps) {
  const router = useRouter();

  useEffect(() => {
    let isMounted = true;

    async function checkSession() {
      const supabase = getSupabaseBrowserClient();
      const {
        data: { session },
      } = await supabase.auth.getSession();

      if (!isMounted) {
        return;
      }

      router.replace(session ? toAuthenticated : toUnauthenticated);
    }

    void checkSession();

    return () => {
      isMounted = false;
    };
  }, [router, toAuthenticated, toUnauthenticated]);

  return (
    <main className="flex min-h-screen items-center justify-center bg-[#f7f1e7] px-6">
      <div className="rounded-[1.75rem] border border-black/10 bg-white px-6 py-5 text-sm text-black/65 shadow-[0_18px_50px_rgba(0,0,0,0.08)]">
        Preparing your workspace...
      </div>
    </main>
  );
}
