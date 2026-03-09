import { redirect } from "next/navigation";
import { getServerSession } from "next-auth/next";

import { SignInForm } from "@/components/auth/sign-in-form";
import { authOptions } from "@/lib/auth/auth-options";

export default async function SignInPage() {
  const session = await getServerSession(authOptions);
  if (session?.user?.id) {
    redirect("/organizations");
  }

  return <SignInForm enableDevCredentials={process.env.AUTH_ENABLE_DEV_CREDENTIALS === "true"} />;
}
