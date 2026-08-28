import { AppHeader } from "@/components/app-shell/app-header";
import { UpdateDetailWorkspace } from "@/components/updates/update-detail-workspace";
import { getApiBaseUrl } from "@/lib/api";

export default async function UpdatePage({ params }: PageProps<"/updates/[id]">) {
  const { id } = await params;
  const apiBaseUrl = getApiBaseUrl();

  return (
    <main className="min-h-screen bg-stone-50 text-zinc-950">
      <AppHeader />
      <UpdateDetailWorkspace
        endpoint={`${apiBaseUrl}/api/updates/${id}`}
        commentEndpoint={`${apiBaseUrl}/api/comments`}
        likeEndpoint={`${apiBaseUrl}/api/like`}
        unlikeEndpoint={`${apiBaseUrl}/api/unlike`}
      />
    </main>
  );
}
