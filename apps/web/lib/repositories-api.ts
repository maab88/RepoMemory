import {
  JobsService,
  RepositoriesService,
  SearchService,
  type GenerateMemoryData,
  type GenerateDigestData,
  type JobResponseData,
  type DigestListData,
  type MemorySearchData,
  type MemoryEntryDetailData,
  type MemoryEntryListData,
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

export function listRepositoryMemory(repoId: string): Promise<MemoryEntryListData> {
  return unwrapData(RepositoriesService.listRepositoryMemory(repoId));
}

export function getRepositoryMemoryDetail(repoId: string, memoryId: string): Promise<MemoryEntryDetailData> {
  return unwrapData(RepositoriesService.getRepositoryMemoryDetail(repoId, memoryId));
}

export function generateRepositoryMemory(repoId: string): Promise<GenerateMemoryData> {
  return unwrapData(RepositoriesService.generateRepositoryMemory(repoId));
}

export function listRepositoryDigests(repoId: string): Promise<DigestListData> {
  return unwrapData(RepositoriesService.listRepositoryDigests(repoId));
}

export function generateRepositoryDigest(repoId: string): Promise<GenerateDigestData> {
  return unwrapData(RepositoriesService.generateRepositoryDigest(repoId));
}

export function searchMemory(params: {
  q?: string;
  organizationId: string;
  repositoryId?: string;
  page?: number;
  pageSize?: number;
}): Promise<MemorySearchData> {
  return unwrapData(
    SearchService.searchMemory(params.organizationId, params.q, params.repositoryId, params.page, params.pageSize)
  );
}
