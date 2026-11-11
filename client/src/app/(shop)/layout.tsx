import { Header } from "@/components/layout/header";
import { NavBar } from "@/components/layout/nav-bar";
import { Footer } from "@/components/layout/footer";

export default function ShopLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col">
      <Header />
      <NavBar />
      <main className="flex-1">{children}</main>
      <Footer />
    </div>
  );
}
