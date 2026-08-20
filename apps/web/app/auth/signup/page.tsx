import Link from "next/link";
import { getApiBaseUrl } from "@/lib/api";

export default function SignUpPage() {
  const apiBaseUrl = getApiBaseUrl();

  return (
    <main className="flex min-h-screen items-center justify-center bg-stone-50 px-6 py-10 text-zinc-950">
      <section className="w-full max-w-md rounded-lg border border-zinc-200 bg-white p-6">
        <Link href="/" className="text-sm font-medium text-sky-700">
          goals
        </Link>
        <h1 className="mt-6 text-2xl font-semibold">Create account</h1>
        <form action={`${apiBaseUrl}/auth/signup`} method="POST" className="mt-6 space-y-4">
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
            <label htmlFor="username" className="block text-sm font-medium">
              Username
            </label>
            <input
              id="username"
              name="username"
              type="text"
              autoComplete="username"
              required
              className="mt-2 w-full rounded-md border border-zinc-300 px-3 py-2 outline-none focus:border-sky-500"
            />
          </div>
          <div>
            <label htmlFor="display_name" className="block text-sm font-medium">
              Display name
            </label>
            <input
              id="display_name"
              name="display_name"
              type="text"
              autoComplete="name"
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
              autoComplete="new-password"
              required
              className="mt-2 w-full rounded-md border border-zinc-300 px-3 py-2 outline-none focus:border-sky-500"
            />
          </div>
          <div>
            <label htmlFor="confirm_password" className="block text-sm font-medium">
              Confirm password
            </label>
            <input
              id="confirm_password"
              name="confirm_password"
              type="password"
              autoComplete="new-password"
              required
              className="mt-2 w-full rounded-md border border-zinc-300 px-3 py-2 outline-none focus:border-sky-500"
            />
          </div>
          <button
            type="submit"
            className="w-full rounded-md bg-zinc-950 px-4 py-2 font-medium text-white hover:bg-zinc-800"
          >
            Create account
          </button>
        </form>
        <p className="mt-4 text-center text-sm text-zinc-600">
          Already have an account?{" "}
          <Link href="/auth/signin" className="font-medium text-sky-700">
            Sign in
          </Link>
        </p>
      </section>
    </main>
  );
}
