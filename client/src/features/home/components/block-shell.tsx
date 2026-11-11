"use client";

import type { ReactNode } from "react";
import type { Block, LayoutEntry } from "@/types/api";
import { ErrorState } from "@/components/ui/error-state";
import { useBlock } from "../hooks/use-block";

type Props<D> = {
  entry: LayoutEntry;
  fail: boolean;
  skeleton: ReactNode;
  children: (block: Block<D>) => ReactNode;
};

// Loads one block and shows its skeleton, its error, or its content.
// Every block has its own query, so one failing block never touches the others.
export function BlockShell<D>({ entry, fail, skeleton, children }: Props<D>) {
  const query = useBlock<D>(entry.id, fail);

  // After Retry the query keeps its error until the new answer comes, so
  // show the skeleton while it runs instead of an error that looks stuck.
  if (query.isPending || (query.isError && query.isFetching)) return skeleton;

  if (query.isError) {
    return (
      <BlockError
        title={entry.title}
        onRetry={() => {
          void query.refetch();
        }}
      />
    );
  }
  return children(query.data);
}

export function BlockError({ title, onRetry }: { title: string; onRetry: () => void }) {
  return (
    <div className="mx-auto max-w-page px-4">
      <ErrorState message={`${title || "This section"} couldn't load.`} onRetry={onRetry} />
    </div>
  );
}
