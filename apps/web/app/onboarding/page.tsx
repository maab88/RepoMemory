"use client";

import { CreateOrganizationForm } from "@/components/organizations/create-organization-form";

export default function OnboardingPage() {
  return (
    <section className="mx-auto max-w-2xl space-y-6">
      <div className="space-y-1 text-center">
        <h2 className="text-3xl font-semibold tracking-tight">Let’s set up your workspace</h2>
        <p className="text-slate-600">Start by creating an organization for your engineering team.</p>
      </div>
      <CreateOrganizationForm />
    </section>
  );
}