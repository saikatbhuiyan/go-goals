type WorkflowCardProps = {
  title: string;
  description: string;
  href: string;
};

export function WorkflowCard({ title, description, href }: WorkflowCardProps) {
  return (
    <a
      href={href}
      className="rounded-lg border border-zinc-200 p-4 transition hover:border-sky-300 hover:bg-sky-50"
    >
      <h2 className="font-semibold">{title}</h2>
      <p className="mt-2 text-sm leading-6 text-zinc-600">{description}</p>
    </a>
  );
}
