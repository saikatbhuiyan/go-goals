type StatusBadgeProps = {
  online: boolean;
};

export function StatusBadge({ online }: StatusBadgeProps) {
  return (
    <span
      className={`rounded-md px-3 py-2 text-sm font-medium ${
        online ? "bg-emerald-50 text-emerald-700" : "bg-amber-50 text-amber-800"
      }`}
    >
      API {online ? "online" : "offline"}
    </span>
  );
}
