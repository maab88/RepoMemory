import {
  JobsService,
  RepositoriesService,
  type JobResponseData,
  type RepositoryListData,
  type OrganizationRepositoriesData,
  type RepositoryDetailData,
  type TriggerSyncData,
} from "@repomemory/contracts";

import { unwrapData } from "@/lib/api-client";

export function listOrganizationRepositories(orgId: string): Promise<OrganizationRepositoriesData> {
  return unwrapData(RepositoriesService.listOrganizationRepositories(orgId));
}

export function listRepositories(): Promise<RepositoryListData> {
  return unwrapData(RepositoriesService.listRepositories());
}

export function getRepositoryDetail(repoId: string): Promise<RepositoryDetailData> {
  return unwrapData(RepositoriesService.getRepositoryDetail(repoId));
}

export function triggerRepositorySync(repoId: string): Promise<TriggerSyncData> {
  return unwrapData(RepositoriesService.triggerRepositorySync(repoId));
}

export function getJob(jobId: string): Promise<JobResponseData> {
  return unwrapData(JobsService.getJob(jobId));
}
