/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { OrganizationRepositoriesResponse } from '../models/OrganizationRepositoriesResponse';
import type { RepositoryDetailResponse } from '../models/RepositoryDetailResponse';
import type { RepositoryListResponse } from '../models/RepositoryListResponse';
import type { TriggerSyncResponse } from '../models/TriggerSyncResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class RepositoriesService {
    /**
     * List imported repositories for an organization
     * @param orgId Organization identifier
     * @returns OrganizationRepositoriesResponse Imported repositories for organization
     * @throws ApiError
     */
    public static listOrganizationRepositories(
        orgId: string,
    ): CancelablePromise<OrganizationRepositoriesResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/organizations/{orgId}/repositories',
            path: {
                'orgId': orgId,
            },
            errors: {
                400: `Invalid request payload or parameters`,
                401: `Missing or invalid auth headers`,
                403: `Authenticated user is not allowed to access this resource`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * List imported repositories for current user
     * @returns RepositoryListResponse Imported repositories
     * @throws ApiError
     */
    public static listRepositories(): CancelablePromise<RepositoryListResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/repositories',
            errors: {
                401: `Missing or invalid auth headers`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Get one imported repository detail
     * @param repoId Repository identifier
     * @returns RepositoryDetailResponse Repository detail
     * @throws ApiError
     */
    public static getRepositoryDetail(
        repoId: string,
    ): CancelablePromise<RepositoryDetailResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/repositories/{repoId}',
            path: {
                'repoId': repoId,
            },
            errors: {
                400: `Invalid request payload or parameters`,
                401: `Missing or invalid auth headers`,
                403: `Authenticated user is not allowed to access this resource`,
                404: `Resource not found`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Enqueue repository initial sync
     * @param repoId Repository identifier
     * @returns TriggerSyncResponse Sync job queued
     * @throws ApiError
     */
    public static triggerRepositorySync(
        repoId: string,
    ): CancelablePromise<TriggerSyncResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/repositories/{repoId}/sync',
            path: {
                'repoId': repoId,
            },
            errors: {
                400: `Invalid request payload or parameters`,
                401: `Missing or invalid auth headers`,
                403: `Authenticated user is not allowed to access this resource`,
                404: `Resource not found`,
                500: `Internal server error`,
            },
        });
    }
}
