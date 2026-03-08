/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { MeResponse } from '../models/MeResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class IdentityService {
    /**
     * Get current authenticated user
     * @returns MeResponse Current user profile
     * @throws ApiError
     */
    public static getMe(): CancelablePromise<MeResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/me',
            errors: {
                401: `Missing or invalid auth headers`,
            },
        });
    }
}
