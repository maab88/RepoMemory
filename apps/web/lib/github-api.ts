import {
  GitHubService,
  type GitHubCallbackSuccessData,
  type GitHubRepositoriesListData,
  type ImportGitHubRepositoriesData,
  type ImportGitHubRepositoriesRequest,
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

export function listGitHubRepositories(): Promise<GitHubRepositoriesListData> {
  return unwrapData(GitHubService.listGitHubRepositories());
}

export function importGitHubRepositories(input: ImportGitHubRepositoriesRequest): Promise<ImportGitHubRepositoriesData> {
  return unwrapData(GitHubService.importGitHubRepositories(input));
}