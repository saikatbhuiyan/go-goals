const tiers = [
  ["UI tier", "Next.js App Router in apps/web"],
  ["API tier", "Go modular monolith in apps/api"],
  ["Data tier", "PostgreSQL with migrations"],
] as const;

export function SystemBoundary() {
  return (
    <aside className="rounded-lg border border-zinc-200 bg-white p-6">
      <h2 className="font-semibold">System boundary</h2>
      <dl className="mt-4 space-y-4 text-sm">
        {tiers.map(([label, value]) => (
          <div key={label}>
            <dt className="font-medium text-zinc-500">{label}</dt>
            <dd>{value}</dd>
          </div>
        ))}
      </dl>
    </aside>
  );
}
