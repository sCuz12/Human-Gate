"use client";

import type { Session } from "@supabase/supabase-js";

const DEFAULT_TIMEOUT_MS = 15_000;

type APIRequestInit = RequestInit & {
  timeoutMs?: number;
};

export async function apiFetch<T>(
  path: string,
  session: Session | null,
  init?: APIRequestInit,
): Promise<T> {
  const apiURL = process.env.NEXT_PUBLIC_API_URL;
  if (!apiURL) {
    throw new Error("Missing NEXT_PUBLIC_API_URL.");
  }

  const { timeoutMs = DEFAULT_TIMEOUT_MS, signal, ...requestInit } = init ?? {};
  const controller = new AbortController();
  const timeoutID = window.setTimeout(() => controller.abort(), timeoutMs);

  signal?.addEventListener("abort", () => controller.abort(), { once: true });

  const headers = new Headers(init?.headers);
  headers.set("Content-Type", "application/json");

  if (session?.access_token) {
    headers.set("Authorization", `Bearer ${session.access_token}`);
  }

  try {
    const response = await fetch(`${apiURL}${path}`, {
      ...requestInit,
      headers,
      signal: controller.signal,
    });

    if (!response.ok) {
      const payload = (await response.json().catch(() => null)) as
        | { error?: { message?: string } }
        | null;
      throw new Error(payload?.error?.message ?? "Request failed.");
    }

    return (await response.json()) as T;
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") {
      throw new Error("Request timed out. Check the API logs.");
    }
    throw error;
  } finally {
    window.clearTimeout(timeoutID);
  }
}
