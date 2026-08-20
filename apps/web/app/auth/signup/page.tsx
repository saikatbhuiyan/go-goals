import Link from "next/link";
import { AuthCard } from "@/components/auth/auth-card";
import { SignUpForm } from "@/components/auth/sign-up-form";
import { getApiBaseUrl } from "@/lib/api";

export default function SignUpPage() {
  const apiBaseUrl = getApiBaseUrl();

  return (
    <AuthCard
      title="Create account"
      footer={
        <>
          Already have an account?{" "}
          <Link href="/auth/signin" className="font-medium text-sky-700">
            Sign in
          </Link>
        </>
      }
    >
      <SignUpForm action={`${apiBaseUrl}/auth/signup`} />
    </AuthCard>
  );
}
