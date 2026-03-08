/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { CreateOrganizationRequest } from '../models/CreateOrganizationRequest';
import type { OrganizationListResponse } from '../models/OrganizationListResponse';
import type { OrganizationResponse } from '../models/OrganizationResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class OrganizationsService {
    /**
     * List organizations accessible by current user
     * @returns OrganizationListResponse Organizations for the current user
     * @throws ApiError
     */
    public static listOrganizations(): CancelablePromise<OrganizationListResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/organizations',
            errors: {
                401: `Missing or invalid auth headers`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Create organization and owner membership for current user
     * @param requestBody
     * @returns OrganizationResponse Organization created
     * @throws ApiError
     */
    public static createOrganization(
        requestBody: CreateOrganizationRequest,
    ): CancelablePromise<OrganizationResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/organizations',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Invalid request payload or parameters`,
                401: `Missing or invalid auth headers`,
                409: `Resource conflict`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Get organization detail for current user
     * @param orgId Organization identifier
     * @returns OrganizationResponse Organization detail
     * @throws ApiError
     */
    public static getOrganization(
        orgId: string,
    ): CancelablePromise<OrganizationResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/organizations/{orgId}',
            path: {
                'orgId': orgId,
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
