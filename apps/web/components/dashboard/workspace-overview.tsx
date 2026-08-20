import { WorkflowCard } from "@/components/dashboard/workflow-card";
import { StatusBadge } from "@/components/ui/status-badge";

type Workflow = {
  title: string;
  description: string;
  path: string;
};

type WorkspaceOverviewProps = {
  apiBaseUrl: string;
  apiOnline: boolean;
  workflows: Workflow[];
};

export function WorkspaceOverview({
  apiBaseUrl,
  apiOnline,
  workflows,
}: WorkspaceOverviewProps) {
  return (
    <div className="rounded-lg border border-zinc-200 bg-white p-6">
      <div className="mb-6 flex items-center justify-between gap-4">
        <div>
          <p className="text-sm font-medium text-sky-700">
            Modular monolith dashboard
          </p>
          <h1 className="mt-2 text-3xl font-semibold tracking-normal">
            Community goals workspace
          </h1>
        </div>
        <StatusBadge online={apiOnline} />
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        {workflows.map((workflow) => (
          <WorkflowCard
            key={workflow.title}
            title={workflow.title}
            description={workflow.description}
            href={`${apiBaseUrl}${workflow.path}`}
          />
        ))}
      </div>
    </div>
  );
}
