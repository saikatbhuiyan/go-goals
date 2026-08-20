import { AuthField } from "@/components/auth/auth-field";

type SignInFormProps = {
  action: string;
};

export function SignInForm({ action }: SignInFormProps) {
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
        id="password"
        name="password"
        label="Password"
        type="password"
        autoComplete="current-password"
        required
      />
      <button
        type="submit"
        className="w-full rounded-md bg-zinc-950 px-4 py-2 font-medium text-white hover:bg-zinc-800"
      >
        Sign in
      </button>
    </form>
  );
}
