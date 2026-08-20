import Link from "next/link";
import { getApiBaseUrl, getApiHealth } from "@/lib/api";

export const dynamic = "force-dynamic";

const workflows = [
  {
    title: "Profile",
    description: "Edit identity, aspirations, bio, and public links.",
    path: "/profile",
  },
  {
    title: "Browse",
    description: "Scan recent members and the latest community updates.",
    path: "/browse",
  },
  {
    title: "Updates",
    description: "Share progress, gather comments, and cheer others on.",
    path: "/profile",
  },
];

export default async function Home() {
  const health = await getApiHealth();
  const apiBaseUrl = getApiBaseUrl();
  const apiOnline = health?.status === "ok";

  return (
    <main className="min-h-screen bg-stone-50 text-zinc-950">
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

      <section className="mx-auto grid max-w-6xl gap-6 px-6 py-8 lg:grid-cols-[1.5fr_1fr]">
        <div className="rounded-lg border border-zinc-200 bg-white p-6">
          <div className="mb-6 flex items-center justify-between gap-4">
            <div>
              <p className="text-sm font-medium text-sky-700">
                Modular monolith dashboard
              </p>
              <h1 className="mt-2 text-3xl font-semibold tracking-normal">
                Community goals workspace
              </h1>
            </div>
            <span
              className={`rounded-md px-3 py-2 text-sm font-medium ${
                apiOnline
                  ? "bg-emerald-50 text-emerald-700"
                  : "bg-amber-50 text-amber-800"
              }`}
            >
              API {apiOnline ? "online" : "offline"}
            </span>
          </div>

          <div className="grid gap-4 md:grid-cols-3">
            {workflows.map((workflow) => (
              <a
                key={workflow.title}
                href={`${apiBaseUrl}${workflow.path}`}
                className="rounded-lg border border-zinc-200 p-4 transition hover:border-sky-300 hover:bg-sky-50"
              >
                <h2 className="font-semibold">{workflow.title}</h2>
                <p className="mt-2 text-sm leading-6 text-zinc-600">
                  {workflow.description}
                </p>
              </a>
            ))}
          </div>
        </div>

        <aside className="rounded-lg border border-zinc-200 bg-white p-6">
          <h2 className="font-semibold">System boundary</h2>
          <dl className="mt-4 space-y-4 text-sm">
            <div>
              <dt className="font-medium text-zinc-500">UI tier</dt>
              <dd>Next.js App Router in apps/web</dd>
            </div>
            <div>
              <dt className="font-medium text-zinc-500">API tier</dt>
              <dd>Go modular monolith in apps/api</dd>
            </div>
            <div>
              <dt className="font-medium text-zinc-500">Data tier</dt>
              <dd>PostgreSQL with migrations</dd>
            </div>
          </dl>
        </aside>
      </section>
    </main>
  );
}
