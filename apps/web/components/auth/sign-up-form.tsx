"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import { AuthField } from "@/components/auth/auth-field";

type SignUpFormProps = {
  endpoint: string;
};

type ApiError = {
  error?: string;
};

export function SignUpForm({ endpoint }: SignUpFormProps) {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setPending(true);

    const formData = new FormData(event.currentTarget);

    try {
      const response = await fetch(endpoint, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          email: String(formData.get("email") ?? ""),
          username: String(formData.get("username") ?? ""),
          display_name: String(formData.get("display_name") ?? ""),
          password: String(formData.get("password") ?? ""),
          confirm_password: String(formData.get("confirm_password") ?? ""),
        }),
      });

      if (!response.ok) {
        const payload = (await response.json().catch(() => null)) as ApiError | null;
        throw new Error(payload?.error ?? "Unable to create account");
      }

      router.push("/");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to create account");
    } finally {
      setPending(false);
    }
  }

  return (
    <form onSubmit={onSubmit} className="mt-6 space-y-4">
      <AuthField
        id="email"
        name="email"
        label="Email"
        type="email"
        autoComplete="email"
        required
      />
      <AuthField
        id="username"
        name="username"
        label="Username"
        type="text"
        autoComplete="username"
        required
      />
      <AuthField
        id="display_name"
        name="display_name"
        label="Display name"
        type="text"
        autoComplete="name"
      />
      <AuthField
        id="password"
        name="password"
        label="Password"
        type="password"
        autoComplete="new-password"
        required
      />
      <AuthField
        id="confirm_password"
        name="confirm_password"
        label="Confirm password"
        type="password"
        autoComplete="new-password"
        required
      />
      {error ? (
        <p className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </p>
      ) : null}
      <button
        type="submit"
        disabled={pending}
        className="w-full rounded-md bg-zinc-950 px-4 py-2 font-medium text-white hover:bg-zinc-800"
      >
        {pending ? "Creating account..." : "Create account"}
      </button>
    </form>
  );
}
