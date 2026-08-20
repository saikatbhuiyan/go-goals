import { AuthField } from "@/components/auth/auth-field";

type SignUpFormProps = {
  action: string;
};

export function SignUpForm({ action }: SignUpFormProps) {
  return (
    <form action={action} method="POST" className="mt-6 space-y-4">
      <AuthField
        id="email"
        name="email"
        label="Email"
        type="email"
        autoComplete="email"
        required
      />
      <AuthField
        id="username"
        name="username"
        label="Username"
        type="text"
        autoComplete="username"
        required
      />
      <AuthField
        id="display_name"
        name="display_name"
        label="Display name"
        type="text"
        autoComplete="name"
      />
      <AuthField
        id="password"
        name="password"
        label="Password"
        type="password"
        autoComplete="new-password"
        required
      />
      <AuthField
        id="confirm_password"
        name="confirm_password"
        label="Confirm password"
        type="password"
        autoComplete="new-password"
        required
      />
      <button
        type="submit"
        className="w-full rounded-md bg-zinc-950 px-4 py-2 font-medium text-white hover:bg-zinc-800"
      >
        Create account
      </button>
    </form>
  );
}
