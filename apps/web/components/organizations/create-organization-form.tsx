"use client";

import { useMutation } from "@tanstack/react-query";
import { FormEvent, useState } from "react";

import { createOrganization } from "@/lib/organizations-api";

export function CreateOrganizationForm() {
  const [name, setName] = useState("");
  const [validationError, setValidationError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: createOrganization,
    onSuccess: (org) => {
      window.location.assign(`/organizations/${org.id}`);
    },
  });

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const trimmed = name.trim();
    if (trimmed.length < 2 || trimmed.length > 80) {
      setValidationError("Organization name must be between 2 and 80 characters.");
      return;
    }

    setValidationError(null);
    mutation.mutate({ name: trimmed });
  };

  return (
    <form onSubmit={onSubmit} className="space-y-4 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="space-y-1">
        <h2 className="text-xl font-semibold">Create your organization</h2>
        <p className="text-sm text-slate-600">You can add repositories after this step.</p>
      </div>

      <div className="space-y-2">
        <label htmlFor="org-name" className="block text-sm font-medium text-slate-700">
          Organization name
        </label>
        <input
          id="org-name"
          name="name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder="Acme Engineering"
          className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm shadow-sm outline-none ring-slate-900/10 transition focus:border-slate-900 focus:ring"
        />
      </div>

      {validationError ? <p className="text-sm text-rose-600">{validationError}</p> : null}
      {mutation.error ? <p className="text-sm text-rose-600">Could not create organization. Try again.</p> : null}

      <button
        type="submit"
        disabled={mutation.isPending}
        className="inline-flex rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
      >
        {mutation.isPending ? "Creating..." : "Create organization"}
      </button>
    </form>
  );
}