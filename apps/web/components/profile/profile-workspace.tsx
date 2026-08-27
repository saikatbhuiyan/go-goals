"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";
import Link from "next/link";

type NullableText = {
  String: string;
  Valid: boolean;
};

type ProfileUser = {
  email: string;
  username: string;
  display_name: string;
  life_aspirations: NullableText;
  things_i_like_to_do: NullableText;
  bio: NullableText;
  bio_link: NullableText;
};

type ProfileResponse = {
  user: ProfileUser;
  follower_count: number;
  updates: Array<{
    id: number;
    content: string;
    created_at: string;
    like_count: number;
    comment_count: number;
  }>;
};

type ProfileWorkspaceProps = {
  profileEndpoint: string;
  editEndpoint: string;
};

type FormState = {
  username: string;
  displayName: string;
  lifeAspirations: string;
  thingsILikeToDo: string;
  bio: string;
  bioLink: string;
};

function readNullable(value: NullableText | undefined) {
  return value?.Valid ? value.String : "";
}

function toFormState(profile: ProfileResponse): FormState {
  return {
    username: profile.user.username,
    displayName: profile.user.display_name,
    lifeAspirations: readNullable(profile.user.life_aspirations),
    thingsILikeToDo: readNullable(profile.user.things_i_like_to_do),
    bio: readNullable(profile.user.bio),
    bioLink: readNullable(profile.user.bio_link),
  };
}

export function ProfileWorkspace({
  profileEndpoint,
  editEndpoint,
}: ProfileWorkspaceProps) {
  const [profile, setProfile] = useState<ProfileResponse | null>(null);
  const [form, setForm] = useState<FormState | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadProfile = useCallback(async () => {
    try {
      const response = await fetch(profileEndpoint, {
        credentials: "include",
        cache: "no-store",
      });

      if (!response.ok) {
        throw new Error(response.status === 401 ? "Sign in required" : "Unable to load profile");
      }

      const payload = (await response.json()) as ProfileResponse;
      setProfile(payload);
      setForm(toFormState(payload));
      setError(null);
    } catch (err) {
      setProfile(null);
      setForm(null);
      setError(err instanceof Error ? err.message : "Unable to load profile");
    } finally {
      setLoading(false);
    }
  }, [profileEndpoint]);

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      void loadProfile();
    }, 0);

    return () => window.clearTimeout(timeout);
  }, [loadProfile]);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!form) {
      return;
    }

    setSaving(true);
    setMessage(null);
    setError(null);

    try {
      const response = await fetch(editEndpoint, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          username: form.username,
          display_name: form.displayName,
          life_aspirations: form.lifeAspirations,
          things_i_like_to_do: form.thingsILikeToDo,
          bio: form.bio,
          bio_link: form.bioLink,
        }),
      });

      if (!response.ok) {
        const payload = (await response.json().catch(() => null)) as { error?: string } | null;
        throw new Error(payload?.error ?? "Unable to save profile");
      }

      setMessage("Profile saved");
      await loadProfile();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to save profile");
    } finally {
      setSaving(false);
    }
  }

  return (
    <section className="mx-auto grid max-w-6xl gap-6 px-6 py-8 lg:grid-cols-[1.2fr_0.8fr]">
      <div className="rounded-lg border border-zinc-200 bg-white p-6">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-sm font-medium text-sky-700">Profile</p>
            <h1 className="mt-2 text-3xl font-semibold tracking-normal">
              {profile?.user.display_name || profile?.user.username || "Your workspace"}
            </h1>
          </div>
          <Link
            href="/browse"
            className="rounded-md border border-zinc-200 px-3 py-2 text-sm hover:bg-zinc-50"
          >
            Browse
          </Link>
        </div>

        {loading ? (
          <p className="mt-6 text-sm text-zinc-600">Loading profile...</p>
        ) : error === "Sign in required" ? (
          <div className="mt-6 rounded-lg border border-zinc-200 bg-zinc-50 p-4">
            <p className="text-sm text-zinc-700">Sign in to manage your profile.</p>
            <Link
              href="/auth/signin"
              className="mt-4 inline-flex rounded-md bg-zinc-950 px-3 py-2 text-sm font-medium text-white hover:bg-zinc-800"
            >
              Sign in
            </Link>
          </div>
        ) : form ? (
          <form onSubmit={onSubmit} className="mt-6 grid gap-4">
            <ProfileInput
              label="Username"
              value={form.username}
              onChange={(username) => setForm({ ...form, username })}
              required
            />
            <ProfileInput
              label="Display name"
              value={form.displayName}
              onChange={(displayName) => setForm({ ...form, displayName })}
            />
            <ProfileTextArea
              label="Life aspirations"
              value={form.lifeAspirations}
              onChange={(lifeAspirations) => setForm({ ...form, lifeAspirations })}
            />
            <ProfileTextArea
              label="Things I like to do"
              value={form.thingsILikeToDo}
              onChange={(thingsILikeToDo) => setForm({ ...form, thingsILikeToDo })}
            />
            <ProfileTextArea
              label="Bio"
              value={form.bio}
              onChange={(bio) => setForm({ ...form, bio })}
            />
            <ProfileInput
              label="Bio link"
              value={form.bioLink}
              onChange={(bioLink) => setForm({ ...form, bioLink })}
            />
            {message ? <p className="text-sm text-emerald-700">{message}</p> : null}
            {error ? <p className="text-sm text-red-700">{error}</p> : null}
            <button
              type="submit"
              disabled={saving}
              className="rounded-md bg-zinc-950 px-4 py-2 font-medium text-white hover:bg-zinc-800 disabled:opacity-60"
            >
              {saving ? "Saving..." : "Save profile"}
            </button>
          </form>
        ) : (
          <p className="mt-6 text-sm text-red-700">{error ?? "Unable to load profile"}</p>
        )}
      </div>

      <aside className="rounded-lg border border-zinc-200 bg-white p-6">
        <p className="text-sm font-medium text-sky-700">Activity</p>
        <div className="mt-4 grid grid-cols-2 gap-3">
          <Stat label="Followers" value={profile?.follower_count ?? 0} />
          <Stat label="Updates" value={profile?.updates.length ?? 0} />
        </div>
        <div className="mt-6 space-y-3">
          {profile?.updates.slice(0, 5).map((update) => (
            <article key={update.id} className="rounded-lg border border-zinc-200 p-3">
              <p className="text-sm text-zinc-700">{update.content}</p>
              <p className="mt-2 text-xs text-zinc-500">
                {update.like_count} likes / {update.comment_count} comments
              </p>
            </article>
          ))}
        </div>
      </aside>
    </section>
  );
}

function ProfileInput({
  label,
  value,
  onChange,
  required = false,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  required?: boolean;
}) {
  return (
    <label className="text-sm font-medium">
      {label}
      <input
        value={value}
        required={required}
        onChange={(event) => onChange(event.target.value)}
        className="mt-2 w-full rounded-md border border-zinc-300 px-3 py-2 outline-none focus:border-sky-500"
      />
    </label>
  );
}

function ProfileTextArea({
  label,
  value,
  onChange,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <label className="text-sm font-medium">
      {label}
      <textarea
        value={value}
        onChange={(event) => onChange(event.target.value)}
        rows={3}
        className="mt-2 w-full rounded-md border border-zinc-300 px-3 py-2 outline-none focus:border-sky-500"
      />
    </label>
  );
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg border border-zinc-200 p-3">
      <p className="text-2xl font-semibold">{value}</p>
      <p className="text-xs text-zinc-500">{label}</p>
    </div>
  );
}
