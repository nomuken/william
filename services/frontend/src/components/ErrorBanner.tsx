type ErrorBannerProps = {
  message: string;
};

export function ErrorBanner({ message }: ErrorBannerProps) {
  if (!message) {
    return null;
  }

  return (
    <div className="rounded-lg border border-neutral-700 bg-neutral-900/80 p-4 text-sm text-neutral-200">
      {message}
    </div>
  );
}
