import { TriangleAlert } from "lucide-react";
import { Button } from "./button";

type Props = { message: string; onRetry?: () => void; retryLabel?: string; className?: string };

export function ErrorState({ message, onRetry, retryLabel = "Retry", className = "" }: Props) {
  return (
    <div
      role="alert"
      className={`flex flex-col items-center justify-center gap-4 border border-red-dark bg-brand-darkest px-6 py-10 text-center ${className}`}
    >
      <TriangleAlert className="size-8 text-red-pink" aria-hidden />
      <p className="text-base font-semibold text-text-secondary">{message}</p>
      {onRetry && (
        <Button variant="outline" onClick={onRetry}>
          {retryLabel}
        </Button>
      )}
    </div>
  );
}
