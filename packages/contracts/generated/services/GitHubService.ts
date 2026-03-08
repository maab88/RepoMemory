/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { GitHubCallbackSuccessResponse } from '../models/GitHubCallbackSuccessResponse';
import type { StartGitHubConnectRequest } from '../models/StartGitHubConnectRequest';
import type { StartGitHubConnectResponse } from '../models/StartGitHubConnectResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class GitHubService {
    /**
     * Start GitHub OAuth connection flow
     * @param requestBody
     * @returns StartGitHubConnectResponse OAuth authorize URL ready
     * @throws ApiError
     */
    public static startGitHubConnect(
        requestBody?: StartGitHubConnectRequest,
    ): CancelablePromise<StartGitHubConnectResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/github/connect/start',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Invalid request payload or parameters`,
                401: `Missing or invalid auth headers`,
                403: `Authenticated user is not allowed to access this resource`,
                500: `Internal server error`,
                503: `Service is not configured or temporarily unavailable`,
            },
        });
    }
    /**
     * Complete GitHub OAuth callback and persist account
     * @param code
     * @param state
     * @returns GitHubCallbackSuccessResponse OAuth callback completed
     * @throws ApiError
     */
    public static completeGitHubCallback(
        code: string,
        state: string,
    ): CancelablePromise<GitHubCallbackSuccessResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/github/callback',
            query: {
                'code': code,
                'state': state,
            },
            errors: {
                400: `Invalid request payload or parameters`,
                401: `Missing or invalid auth headers`,
                403: `Authenticated user is not allowed to access this resource`,
                500: `Internal server error`,
                502: `Upstream integration failed`,
                503: `Service is not configured or temporarily unavailable`,
            },
        });
    }
}
