"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";

type RecentUser = {
  id: number;
  username: string;
  display_name: string;
  profile_image_url: string;
};

type FeedUpdate = {
  id: number;
  username: string;
  display_name: string;
  content: string;
  created_at: string;
  like_count: number;
  comment_count: number;
};

type FeedResponse = {
  users: RecentUser[];
  updates: FeedUpdate[];
};

type BrowseWorkspaceProps = {
  endpoint: string;
};

export function BrowseWorkspace({ endpoint }: BrowseWorkspaceProps) {
  const [feed, setFeed] = useState<FeedResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadFeed = useCallback(async () => {
    try {
      const response = await fetch(endpoint, {
        credentials: "include",
        cache: "no-store",
      });

      if (!response.ok) {
        throw new Error("Unable to load browse feed");
      }

      setFeed((await response.json()) as FeedResponse);
      setError(null);
    } catch (err) {
      setFeed(null);
      setError(err instanceof Error ? err.message : "Unable to load browse feed");
    } finally {
      setLoading(false);
    }
  }, [endpoint]);

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      void loadFeed();
    }, 0);

    return () => window.clearTimeout(timeout);
  }, [loadFeed]);

  return (
    <section className="mx-auto grid max-w-6xl gap-6 px-6 py-8 lg:grid-cols-[1.2fr_0.8fr]">
      <div className="rounded-lg border border-zinc-200 bg-white p-6">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-sm font-medium text-sky-700">Browse</p>
            <h1 className="mt-2 text-3xl font-semibold tracking-normal">
              Community updates
            </h1>
          </div>
          <Link
            href="/profile"
            className="rounded-md border border-zinc-200 px-3 py-2 text-sm hover:bg-zinc-50"
          >
            Profile
          </Link>
        </div>

        {loading ? (
          <p className="mt-6 text-sm text-zinc-600">Loading updates...</p>
        ) : error ? (
          <p className="mt-6 text-sm text-red-700">{error}</p>
        ) : (
          <div className="mt-6 space-y-4">
            {feed?.updates.length ? (
              feed.updates.map((update) => (
                <article key={update.id} className="rounded-lg border border-zinc-200 p-4">
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <p className="font-medium">
                        {update.display_name || update.username}
                      </p>
                      <p className="text-sm text-zinc-500">@{update.username}</p>
                    </div>
                    <p className="text-xs text-zinc-500">
                      {update.like_count} likes / {update.comment_count} comments
                    </p>
                  </div>
                  <p className="mt-3 text-sm leading-6 text-zinc-700">{update.content}</p>
                </article>
              ))
            ) : (
              <p className="text-sm text-zinc-600">No updates yet.</p>
            )}
          </div>
        )}
      </div>

      <aside className="rounded-lg border border-zinc-200 bg-white p-6">
        <p className="text-sm font-medium text-sky-700">Members</p>
        <div className="mt-4 space-y-3">
          {feed?.users.length ? (
            feed.users.slice(0, 12).map((user) => (
              <div key={user.id} className="flex items-center gap-3 rounded-lg border border-zinc-200 p-3">
                <div className="grid h-9 w-9 place-items-center rounded-md bg-zinc-100 text-sm font-semibold">
                  {(user.display_name || user.username || "?").slice(0, 1).toUpperCase()}
                </div>
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">
                    {user.display_name || user.username}
                  </p>
                  <p className="truncate text-xs text-zinc-500">@{user.username}</p>
                </div>
              </div>
            ))
          ) : (
            <p className="text-sm text-zinc-600">No members yet.</p>
          )}
        </div>
      </aside>
    </section>
  );
}
