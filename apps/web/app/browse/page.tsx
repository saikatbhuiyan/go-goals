import { AppHeader } from "@/components/app-shell/app-header";
import { BrowseWorkspace } from "@/components/browse/browse-workspace";
import { getApiBaseUrl } from "@/lib/api";

export default function BrowsePage() {
  const apiBaseUrl = getApiBaseUrl();

  return (
    <main className="min-h-screen bg-stone-50 text-zinc-950">
      <AppHeader />
      <BrowseWorkspace endpoint={`${apiBaseUrl}/api/browse`} />
    </main>
  );
}
