import { GitHubCallbackStatus } from "@/components/integrations/github-callback-status";

type GitHubCallbackPageProps = {
  searchParams: Promise<{
    code?: string;
    state?: string;
  }>;
};

export default async function GitHubCallbackPage({ searchParams }: GitHubCallbackPageProps) {
  const { code, state } = await searchParams;

  return <GitHubCallbackStatus code={code} state={state} />;
}