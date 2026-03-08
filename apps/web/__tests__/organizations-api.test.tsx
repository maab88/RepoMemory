import { IdentityService, OrganizationsService } from "@repomemory/contracts";

import { createOrganization, getCurrentUser, getOrganization, listOrganizations } from "@/lib/organizations-api";

describe("organizations-api contract integration", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("uses generated service methods and unwraps data payloads", async () => {
    const user = { id: "u1", displayName: "Dev User" };
    const organization = { id: "o1", name: "Acme", slug: "acme", role: "owner" as const };

    const getMeSpy = vi.spyOn(IdentityService, "getMe").mockResolvedValue({ data: user } as never);
    const listSpy = vi
      .spyOn(OrganizationsService, "listOrganizations")
      .mockResolvedValue({ data: [organization] } as never);
    const createSpy = vi
      .spyOn(OrganizationsService, "createOrganization")
      .mockResolvedValue({ data: organization } as never);
    const getSpy = vi
      .spyOn(OrganizationsService, "getOrganization")
      .mockResolvedValue({ data: organization } as never);

    await expect(getCurrentUser()).resolves.toEqual(user);
    await expect(listOrganizations()).resolves.toEqual([organization]);
    await expect(createOrganization({ name: "Acme" })).resolves.toEqual(organization);
    await expect(getOrganization("o1")).resolves.toEqual(organization);

    expect(getMeSpy).toHaveBeenCalledTimes(1);
    expect(listSpy).toHaveBeenCalledTimes(1);
    expect(createSpy).toHaveBeenCalledWith({ name: "Acme" });
    expect(getSpy).toHaveBeenCalledWith("o1");
  });
});
