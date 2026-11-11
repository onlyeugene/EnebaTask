import { NotFoundView } from "@/components/layout/not-found-view";

// Shown inside the shop layout (header, nav, footer) whenever a page in this
// group calls notFound(): the catch-all route below, and the product page.
export default function NotFound() {
  return <NotFoundView />;
}
