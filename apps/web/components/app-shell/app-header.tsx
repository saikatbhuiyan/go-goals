import Link from "next/link";
import { AuthNav } from "@/components/app-shell/auth-nav";
import { getApiBaseUrl } from "@/lib/api";

export function AppHeader() {
  const apiBaseUrl = getApiBaseUrl();

  return (
    <header className="border-b border-zinc-200 bg-white">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <Link href="/" className="text-xl font-semibold">
          goals
        </Link>
        <AuthNav
          profileEndpoint={`${apiBaseUrl}/api/profile`}
          logoutEndpoint={`${apiBaseUrl}/api/auth/logout`}
        />
      </div>
    </header>
  );
}
