import Link from "next/link";
import { AuthCard } from "@/components/auth/auth-card";
import { SignInForm } from "@/components/auth/sign-in-form";
import { getApiBaseUrl } from "@/lib/api";

export default function SignInPage() {
  const apiBaseUrl = getApiBaseUrl();

  return (
    <AuthCard
      title="Sign in"
      footer={
        <>
          New here?{" "}
          <Link href="/auth/signup" className="font-medium text-sky-700">
            Create an account
          </Link>
        </>
      }
    >
      <SignInForm endpoint={`${apiBaseUrl}/api/auth/signin`} />
    </AuthCard>
  );
}
