import {
  IdentityService,
  OrganizationsService,
  type CreateOrganizationRequest,
  type Organization,
  type User,
} from "@repomemory/contracts";
import { unwrapData } from "@/lib/api-client";

export function getCurrentUser(): Promise<User> {
  return unwrapData(IdentityService.getMe());
}

export function listOrganizations(): Promise<Organization[]> {
  return unwrapData(OrganizationsService.listOrganizations());
}

export function createOrganization(input: CreateOrganizationRequest): Promise<Organization> {
  return unwrapData(OrganizationsService.createOrganization(input));
}

export function getOrganization(orgId: string): Promise<Organization> {
  return unwrapData(OrganizationsService.getOrganization(orgId));
}