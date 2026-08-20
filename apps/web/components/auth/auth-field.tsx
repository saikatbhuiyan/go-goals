type AuthFieldProps = {
  id: string;
  name: string;
  label: string;
  type: "email" | "password" | "text";
  autoComplete: string;
  required?: boolean;
};

export function AuthField({
  id,
  name,
  label,
  type,
  autoComplete,
  required = false,
}: AuthFieldProps) {
  return (
    <div>
      <label htmlFor={id} className="block text-sm font-medium">
        {label}
      </label>
      <input
        id={id}
        name={name}
        type={type}
        autoComplete={autoComplete}
        required={required}
        className="mt-2 w-full rounded-md border border-zinc-300 px-3 py-2 outline-none focus:border-sky-500"
      />
    </div>
  );
}
