type JobFailureBannerProps = {
  message?: string | null;
};

export function JobFailureBanner({ message }: JobFailureBannerProps) {
  return (
    <div className="rounded-lg border border-rose-200 bg-rose-50 p-3 text-rose-800">
      <p className="text-sm font-medium">Background job failed</p>
      <p className="mt-1 text-sm">{message || "The job failed. Please retry the action."}</p>
    </div>
  );
}
