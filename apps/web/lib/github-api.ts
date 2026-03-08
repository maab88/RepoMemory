import {
  GitHubService,
  type GitHubCallbackSuccessData,
  type StartGitHubConnectData,
  type StartGitHubConnectRequest,
} from "@repomemory/contracts";

import { unwrapData } from "@/lib/api-client";

export function startGitHubConnect(input?: StartGitHubConnectRequest): Promise<StartGitHubConnectData> {
  return unwrapData(GitHubService.startGitHubConnect(input));
}

export function completeGitHubCallback(code: string, state: string): Promise<GitHubCallbackSuccessData> {
  return unwrapData(GitHubService.completeGitHubCallback(code, state));
}