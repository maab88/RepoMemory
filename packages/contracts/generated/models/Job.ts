/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { JobPayload } from './JobPayload';
export type Job = {
    id: string;
    jobType: 'repo.initial_sync' | 'repo.incremental_sync' | 'repo.generate_memory' | 'repo.generate_digest' | 'repo.recalculate_hotspots';
    status: 'queued' | 'running' | 'succeeded' | 'failed';
    queueName: string;
    attempts: number;
    lastError?: string | null;
    payload: JobPayload;
    startedAt?: string | null;
    finishedAt?: string | null;
    createdAt: string;
    updatedAt: string;
};

