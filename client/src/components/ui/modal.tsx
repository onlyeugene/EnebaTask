"use client";

import { useEffect, useRef, type ReactNode } from "react";

type Props = { open: boolean; onClose: () => void; label: string; children: ReactNode };

// The browser's own <dialog>: showModal() dims the page, keeps keyboard focus
// inside, and closes on Escape, so none of that is written by hand here.
export function Modal({ open, onClose, label, children }: Props) {
  const ref = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const d = ref.current;
    if (!d) return;
    if (open && !d.open) d.showModal();
    if (!open && d.open) d.close();
  }, [open]);

  return (
    <dialog
      ref={ref}
      aria-label={label}
      // onClose fires for Escape too, so the parent's state stays in step.
      onClose={onClose}
      // A click on the dim area outside the box lands on the <dialog> itself.
      onClick={(e) => e.target === e.currentTarget && onClose()}
      className="m-auto w-[440px] max-w-[calc(100vw-32px)] bg-brand-mid text-white backdrop:bg-ink/70"
    >
      {children}
    </dialog>
  );
}
