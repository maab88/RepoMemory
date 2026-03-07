import { apiRequest } from "@/lib/api-client";
import { CurrentUser, Organization } from "@/lib/types";

export function getCurrentUser(): Promise<CurrentUser> {
  return apiRequest<CurrentUser>("/me", { method: "GET" });
}

export function listOrganizations(): Promise<Organization[]> {
  return apiRequest<Organization[]>("/organizations", { method: "GET" });
}

export function createOrganization(input: { name: string }): Promise<Organization> {
  return apiRequest<Organization>("/organizations", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function getOrganization(orgId: string): Promise<Organization> {
  return apiRequest<Organization>(`/organizations/${orgId}`, { method: "GET" });
}