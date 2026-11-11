import { Header } from "@/components/layout/header";
import { NavBar } from "@/components/layout/nav-bar";
import { Footer } from "@/components/layout/footer";
import { SearchBox } from "@/features/search/components/search-box";
import { CartPopover } from "@/features/cart/components/cart-popover";

export default function ShopLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col">
      <Header search={<SearchBox />} cart={<CartPopover />} />
      <NavBar />
      <main className="flex-1">{children}</main>
      <Footer />
    </div>
  );
}
