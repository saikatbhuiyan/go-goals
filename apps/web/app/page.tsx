import { AppHeader } from "@/components/app-shell/app-header";
import { SystemBoundary } from "@/components/dashboard/system-boundary";
import { WorkspaceOverview } from "@/components/dashboard/workspace-overview";
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
      <AppHeader />

      <section className="mx-auto grid max-w-6xl gap-6 px-6 py-8 lg:grid-cols-[1.5fr_1fr]">
        <WorkspaceOverview
          apiBaseUrl={apiBaseUrl}
          apiOnline={apiOnline}
          workflows={workflows}
        />
        <SystemBoundary />
      </section>
    </main>
  );
}
