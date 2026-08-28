import { AppHeader } from "@/components/app-shell/app-header";
import { ProfileWorkspace } from "@/components/profile/profile-workspace";
import { getApiBaseUrl } from "@/lib/api";

export default function ProfilePage() {
  const apiBaseUrl = getApiBaseUrl();

  return (
    <main className="min-h-screen bg-stone-50 text-zinc-950">
      <AppHeader />
      <ProfileWorkspace
        profileEndpoint={`${apiBaseUrl}/api/profile`}
        editEndpoint={`${apiBaseUrl}/api/profile/edit`}
        updateEndpoint={`${apiBaseUrl}/api/aspiration-updates`}
        deleteUpdateEndpoint={`${apiBaseUrl}/api/aspiration-updates/delete`}
      />
    </main>
  );
}
