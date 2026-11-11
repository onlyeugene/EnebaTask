"use client";

import { useState } from "react";
import { Check, Copy } from "lucide-react";

// One key code with a Copy button. The code is plain text too, so it can be
// selected by hand where the clipboard isn't allowed (some browsers block it
// outside https).
export function KeyReveal({ code }: { code: string }) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard refused: the code is still on screen to select.
    }
  };

  return (
    <div className="flex items-center justify-between gap-3 border border-mint bg-brand-darkest px-4 py-3">
      <code className="text-base font-bold tracking-wider break-all select-all">{code}</code>
      <button
        type="button"
        onClick={copy}
        className="link-hover flex shrink-0 items-center gap-1.5 text-sm font-bold"
        aria-label={copied ? "Copied" : `Copy key ${code}`}
      >
        {copied ? <Check className="size-4 text-mint" aria-hidden /> : <Copy className="size-4" aria-hidden />}
        {copied ? "Copied" : "Copy"}
      </button>
    </div>
  );
}
