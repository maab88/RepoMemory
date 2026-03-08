import type { MeResponse, OrganizationListResponse, OrganizationResponse } from "@repomemory/contracts";
import { createOrganization, getCurrentUser, getOrganization, listOrganizations } from "@/lib/organizations-api";

type IsEqual<A, B> =
  (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;

type _CurrentUserType = Assert<IsEqual<Awaited<ReturnType<typeof getCurrentUser>>, MeResponse["data"]>>;
type _ListOrganizationsType = Assert<IsEqual<Awaited<ReturnType<typeof listOrganizations>>, OrganizationListResponse["data"]>>;
type _CreateOrganizationType = Assert<IsEqual<Awaited<ReturnType<typeof createOrganization>>, OrganizationResponse["data"]>>;
type _GetOrganizationType = Assert<IsEqual<Awaited<ReturnType<typeof getOrganization>>, OrganizationResponse["data"]>>;
