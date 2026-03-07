import { OrganizationDetailClient } from "@/components/organizations/organization-detail";

type OrganizationDetailPageProps = {
  params: Promise<{
    orgId: string;
  }>;
};

export default async function OrganizationDetailPage({ params }: OrganizationDetailPageProps) {
  const { orgId } = await params;
  return <OrganizationDetailClient orgId={orgId} />;
}