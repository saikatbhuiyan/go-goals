"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";

type AuthNavProps = {
  profileEndpoint: string;
  logoutEndpoint: string;
};

export function AuthNav({ profileEndpoint, logoutEndpoint }: AuthNavProps) {
  const router = useRouter();
  const [signedIn, setSignedIn] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const [signingOut, setSigningOut] = useState(false);

  const loadSession = useCallback(async () => {
    try {
      const response = await fetch(profileEndpoint, {
        credentials: "include",
        cache: "no-store",
      });
      setSignedIn(response.ok);
    } catch {
      setSignedIn(false);
    } finally {
      setLoaded(true);
    }
  }, [profileEndpoint]);

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      void loadSession();
    }, 0);

    return () => window.clearTimeout(timeout);
  }, [loadSession]);

  async function signOut() {
    setSigningOut(true);

    try {
      await fetch(logoutEndpoint, {
        method: "POST",
        credentials: "include",
      });
      setSignedIn(false);
      router.push("/");
      router.refresh();
    } finally {
      setSigningOut(false);
    }
  }

  if (!loaded) {
    return <div className="h-9 w-28" />;
  }

  if (signedIn) {
    return (
      <nav className="flex items-center gap-2 text-sm">
        <Link
          href="/browse"
          className="rounded-md px-3 py-2 text-zinc-700 hover:bg-zinc-100"
        >
          Browse
        </Link>
        <Link
          href="/profile"
          className="rounded-md px-3 py-2 text-zinc-700 hover:bg-zinc-100"
        >
          Profile
        </Link>
        <button
          type="button"
          onClick={() => void signOut()}
          disabled={signingOut}
          className="rounded-md bg-zinc-950 px-3 py-2 text-white hover:bg-zinc-800 disabled:opacity-60"
        >
          {signingOut ? "Signing out..." : "Sign out"}
        </button>
      </nav>
    );
  }

  return (
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
  );
}
