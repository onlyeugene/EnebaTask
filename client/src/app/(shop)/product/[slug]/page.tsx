import { ProductView } from "@/features/product/components/product-view";

// The slug comes from the URL only; the data is loaded in the browser by
// ProductView, like every other data view (CLAUDE.md, Working notes).
export default async function ProductPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  return <ProductView slug={slug} />;
}
