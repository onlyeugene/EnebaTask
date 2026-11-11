import Link from "next/link";
import { buttonStyles } from "@/components/ui/button";
import { PixelShapes } from "./pixel-shapes";

// Eneba's 404 ("Highscore! 404"): big heading, one line, an outlined button
// home, over our own floating pixel shapes. Shown for unknown URLs and for
// product slugs the API doesn't know (PLAN.md, "404 page").
// The heading drops to text-3xl below 600px, where 80px wouldn't fit a phone.
export function NotFoundView() {
  return (
    <section className="relative min-h-[640px] overflow-hidden">
      {/* React 19 moves this into <head>. */}
      <title>Page not found | Eugene</title>
      <PixelShapes />
      <div className="relative mx-auto max-w-page px-4 pt-20 pb-40 sm:pt-28">
        <h1 className="text-3xl font-extrabold sm:text-display">Highscore! 404</h1>
        <p className="mt-6 text-xl font-normal sm:mt-10">We&apos;re pretty sure something went horribly wrong</p>
        <Link href="/" className={`${buttonStyles.outline} mt-10 px-14 text-base sm:mt-14`}>
          Continue browsing
        </Link>
      </div>
    </section>
  );
}
