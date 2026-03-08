/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { JobResponse } from '../models/JobResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class JobsService {
    /**
     * Get one background job by id
     * @param jobId Job identifier
     * @returns JobResponse Job record
     * @throws ApiError
     */
    public static getJob(
        jobId: string,
    ): CancelablePromise<JobResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/jobs/{jobId}',
            path: {
                'jobId': jobId,
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
