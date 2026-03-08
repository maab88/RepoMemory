/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { MemorySearchResponse } from '../models/MemorySearchResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class SearchService {
    /**
     * Search memory entries by title and summary
     * @param organizationId Organization identifier
     * @param q Search query
     * @param repositoryId Optional repository identifier within the organization
     * @param page
     * @param pageSize
     * @returns MemorySearchResponse Search results from persisted memory entries
     * @throws ApiError
     */
    public static searchMemory(
        organizationId: string,
        q?: string,
        repositoryId?: string,
        page: number = 1,
        pageSize: number = 20,
    ): CancelablePromise<MemorySearchResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/memory/search',
            query: {
                'q': q,
                'organizationId': organizationId,
                'repositoryId': repositoryId,
                'page': page,
                'pageSize': pageSize,
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
