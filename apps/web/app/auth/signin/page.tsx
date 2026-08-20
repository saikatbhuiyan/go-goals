import Link from "next/link";
import { getApiBaseUrl } from "@/lib/api";

export default function SignInPage() {
  const apiBaseUrl = getApiBaseUrl();

  return (
    <main className="flex min-h-screen items-center justify-center bg-stone-50 px-6 py-10 text-zinc-950">
      <section className="w-full max-w-md rounded-lg border border-zinc-200 bg-white p-6">
        <Link href="/" className="text-sm font-medium text-sky-700">
          goals
        </Link>
        <h1 className="mt-6 text-2xl font-semibold">Sign in</h1>
        <form action={`${apiBaseUrl}/auth/signin`} method="POST" className="mt-6 space-y-4">
          <div>
            <label htmlFor="email" className="block text-sm font-medium">
              Email
            </label>
            <input
              id="email"
              name="email"
              type="email"
              autoComplete="email"
              required
              className="mt-2 w-full rounded-md border border-zinc-300 px-3 py-2 outline-none focus:border-sky-500"
            />
          </div>
          <div>
            <label htmlFor="password" className="block text-sm font-medium">
              Password
            </label>
            <input
              id="password"
              name="password"
              type="password"
              autoComplete="current-password"
              required
              className="mt-2 w-full rounded-md border border-zinc-300 px-3 py-2 outline-none focus:border-sky-500"
            />
          </div>
          <button
            type="submit"
            className="w-full rounded-md bg-zinc-950 px-4 py-2 font-medium text-white hover:bg-zinc-800"
          >
            Sign in
          </button>
        </form>
        <p className="mt-4 text-center text-sm text-zinc-600">
          New here?{" "}
          <Link href="/auth/signup" className="font-medium text-sky-700">
            Create an account
          </Link>
        </p>
      </section>
    </main>
  );
}
