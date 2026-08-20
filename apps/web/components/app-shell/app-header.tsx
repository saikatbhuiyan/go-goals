import Link from "next/link";

export function AppHeader() {
  return (
    <header className="border-b border-zinc-200 bg-white">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <Link href="/" className="text-xl font-semibold">
          goals
        </Link>
        <nav className="flex items-center gap-3 text-sm">
          <Link
            href="/auth/signin"
            className="rounded-md px-3 py-2 text-zinc-700 hover:bg-zinc-100"
          >
            Sign In
          </Link>
          <Link
            href="/auth/signup"
            className="rounded-md bg-zinc-950 px-3 py-2 text-white hover:bg-zinc-800"
          >
            Sign Up
          </Link>
        </nav>
      </div>
    </header>
  );
}
