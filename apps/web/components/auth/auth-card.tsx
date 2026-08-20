import Link from "next/link";
import type { ReactNode } from "react";

type AuthCardProps = {
  title: string;
  children: ReactNode;
  footer: ReactNode;
};

export function AuthCard({ title, children, footer }: AuthCardProps) {
  return (
    <main className="flex min-h-screen items-center justify-center bg-stone-50 px-6 py-10 text-zinc-950">
      <section className="w-full max-w-md rounded-lg border border-zinc-200 bg-white p-6">
        <Link href="/" className="text-sm font-medium text-sky-700">
          goals
        </Link>
        <h1 className="mt-6 text-2xl font-semibold">{title}</h1>
        {children}
        <p className="mt-4 text-center text-sm text-zinc-600">{footer}</p>
      </section>
    </main>
  );
}
