type RetryBannerProps = {
  message: string;
  onRetry: () => void;
  disabled?: boolean;
};

export function RetryBanner({ message, onRetry, disabled = false }: RetryBannerProps) {
  return (
    <div className="rounded-lg border border-amber-200 bg-amber-50 p-3 text-amber-900">
      <p className="text-sm">{message}</p>
      <button
        type="button"
        onClick={onRetry}
        disabled={disabled}
        className="mt-2 inline-flex rounded-md border border-amber-300 px-3 py-1.5 text-sm font-medium hover:border-amber-400 disabled:cursor-not-allowed disabled:opacity-60"
      >
        Retry
      </button>
    </div>
  );
}
