"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";
import Link from "next/link";

type Comment = {
  id: number;
  update_id: number;
  parent_id: {
    Int64: number;
    Valid: boolean;
  };
  content: string;
  created_at: string;
  username: string;
  display_name: string;
  replies: Comment[];
};

type UpdateDetail = {
  id: number;
  username: string;
  display_name: string;
  content: string;
  created_at: string;
  like_count: number;
  comment_count: number;
  liked: boolean;
  is_own_post: boolean;
};

type UpdateResponse = {
  update: UpdateDetail;
  comments: Comment[];
  total_comments: number;
  is_authenticated: boolean;
};

type UpdateDetailWorkspaceProps = {
  endpoint: string;
  commentEndpoint: string;
  likeEndpoint: string;
  unlikeEndpoint: string;
};

export function UpdateDetailWorkspace({
  endpoint,
  commentEndpoint,
  likeEndpoint,
  unlikeEndpoint,
}: UpdateDetailWorkspaceProps) {
  const [data, setData] = useState<UpdateResponse | null>(null);
  const [comment, setComment] = useState("");
  const [loading, setLoading] = useState(true);
  const [busyAction, setBusyAction] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const comments = data?.comments ?? [];

  const loadUpdate = useCallback(async () => {
    try {
      const response = await fetch(endpoint, {
        credentials: "include",
        cache: "no-store",
      });

      if (!response.ok) {
        const payload = (await response.json().catch(() => null)) as { error?: string } | null;
        throw new Error(payload?.error ?? "Unable to load update");
      }

      setData((await response.json()) as UpdateResponse);
      setError(null);
    } catch (err) {
      setData(null);
      setError(err instanceof Error ? err.message : "Unable to load update");
    } finally {
      setLoading(false);
    }
  }, [endpoint]);

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      void loadUpdate();
    }, 0);

    return () => window.clearTimeout(timeout);
  }, [loadUpdate]);

  async function postJSON(endpointURL: string, body: object) {
    const response = await fetch(endpointURL, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const payload = (await response.json().catch(() => null)) as { error?: string } | null;
      throw new Error(payload?.error ?? "Action failed");
    }
  }

  async function toggleLike() {
    if (!data) {
      return;
    }

    setBusyAction("like");
    setError(null);

    try {
      await postJSON(data.update.liked ? unlikeEndpoint : likeEndpoint, {
        update_id: String(data.update.id),
      });
      await loadUpdate();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to update like");
    } finally {
      setBusyAction(null);
    }
  }

  async function addComment(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!data) {
      return;
    }

    const content = comment.trim();
    if (!content) {
      setError("Comment content cannot be empty");
      return;
    }

    setBusyAction("comment");
    setError(null);

    try {
      await postJSON(commentEndpoint, {
        update_id: String(data.update.id),
        content,
      });
      setComment("");
      await loadUpdate();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to add comment");
    } finally {
      setBusyAction(null);
    }
  }

  return (
    <section className="mx-auto max-w-4xl px-6 py-8">
      <Link href="/browse" className="text-sm font-medium text-sky-700 hover:text-sky-800">
        Back to browse
      </Link>

      {loading ? (
        <div className="mt-6 rounded-lg border border-zinc-200 bg-white p-6">
          <p className="text-sm text-zinc-600">Loading update...</p>
        </div>
      ) : error && !data ? (
        <div className="mt-6 rounded-lg border border-red-200 bg-red-50 p-6">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      ) : data ? (
        <div className="mt-6 space-y-6">
          <article className="rounded-lg border border-zinc-200 bg-white p-6">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="font-medium">
                  {data.update.display_name || data.update.username}
                </p>
                <p className="text-sm text-zinc-500">@{data.update.username}</p>
              </div>
              {!data.update.is_own_post ? (
                <button
                  type="button"
                  onClick={() => void toggleLike()}
                  disabled={busyAction === "like"}
                  className="rounded-md border border-zinc-200 px-3 py-1.5 text-sm font-medium hover:bg-zinc-50 disabled:opacity-60"
                >
                  {busyAction === "like" ? "Saving..." : data.update.liked ? "Unlike" : "Like"}
                </button>
              ) : null}
            </div>
            <p className="mt-4 text-base leading-7 text-zinc-800">{data.update.content}</p>
            <p className="mt-4 text-xs text-zinc-500">
              {data.update.like_count} likes / {data.total_comments} comments
            </p>
          </article>

          {error ? (
            <p className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {error}
            </p>
          ) : null}

          {data.is_authenticated ? (
            <form onSubmit={addComment} className="rounded-lg border border-zinc-200 bg-white p-6">
              <label className="text-sm font-medium">
                Comment
                <textarea
                  value={comment}
                  onChange={(event) => setComment(event.target.value)}
                  rows={4}
                  className="mt-2 w-full rounded-md border border-zinc-300 px-3 py-2 outline-none focus:border-sky-500"
                />
              </label>
              <button
                type="submit"
                disabled={busyAction === "comment"}
                className="mt-3 rounded-md bg-zinc-950 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-800 disabled:opacity-60"
              >
                {busyAction === "comment" ? "Posting..." : "Post comment"}
              </button>
            </form>
          ) : (
            <div className="rounded-lg border border-zinc-200 bg-white p-6">
              <p className="text-sm text-zinc-700">Sign in to comment on this update.</p>
              <Link
                href="/auth/signin"
                className="mt-4 inline-flex rounded-md bg-zinc-950 px-3 py-2 text-sm font-medium text-white hover:bg-zinc-800"
              >
                Sign in
              </Link>
            </div>
          )}

          <div className="rounded-lg border border-zinc-200 bg-white p-6">
            <p className="text-sm font-medium text-sky-700">Comments</p>
            <div className="mt-4 space-y-4">
              {comments.length ? (
                comments.map((item) => <CommentItem key={item.id} comment={item} />)
              ) : (
                <p className="text-sm text-zinc-600">No comments yet.</p>
              )}
            </div>
          </div>
        </div>
      ) : null}
    </section>
  );
}

function CommentItem({ comment }: { comment: Comment }) {
  const replies = comment.replies ?? [];

  return (
    <article className="rounded-lg border border-zinc-200 p-4">
      <p className="text-sm font-medium">
        {comment.display_name || comment.username}
      </p>
      <p className="text-xs text-zinc-500">@{comment.username}</p>
      <p className="mt-3 text-sm leading-6 text-zinc-700">{comment.content}</p>
      {replies.length ? (
        <div className="mt-4 space-y-3 border-l border-zinc-200 pl-4">
          {replies.map((reply) => (
            <CommentItem key={reply.id} comment={reply} />
          ))}
        </div>
      ) : null}
    </article>
  );
}
