// Types from API.md. Types only, no runtime code. Money is integer cents (EUR).

export type User = { id: string; email: string; name: string; createdAt: string };

export type TokenPair = { accessToken: string; refreshToken: string };
export type AuthResponse = TokenPair & { user: User };

export type ErrorBody = {
  error: { code: string; message: string; fields?: Record<string, string> };
};

export type ProductType = "game" | "gift_card" | "software";
export type Platform = "steam" | "psn" | "xbox" | "nintendo" | "windows";
export type Region = "global" | "europe" | "united_states" | "united_kingdom";

export type ProductCard = {
  id: string;
  slug: string;
  name: string;
  type: ProductType;
  platform: Platform;
  region: Region;
  coverUrl: string; // relative to the API URL
  giftCardValue?: string;
  originalPriceCents: number;
  minPriceCents: number | null; // null = sold out
  discountPct: number;
  cashbackPct: number;
  offerCount: number;
};

export type ProductDetail = ProductCard & {
  description: string;
  genres: string[];
  developer: string;
  publisher: string;
  releaseDate: string;
  activation: string;
};

export type Offer = {
  id: string;
  productId: string;
  priceCents: number;
  stock: number;
  seller: { id: string; slug: string; name: string; rating: number; salesCount: number };
};

// GET /api/products/{slug}. Offers come sorted by price, in-stock first.
export type ProductPage = { product: ProductDetail; offers: Offer[] };

export type ProductList ={ items: ProductCard[]; page: number; pageSize: number; total: number };

export type FilterOption = { value: string; label: string; count: number };
export type Filters = {
  types: FilterOption[];
  platforms: FilterOption[];
  regions: FilterOption[];
  genres: FilterOption[];
  price: { minCents: number; maxCents: number };
};

export type Suggestion = {
  slug: string;
  name: string;
  type: ProductType;
  platform: Platform;
  region: Region;
  originalPriceCents: number;
  minPriceCents: number | null;
  discountPct: number;
  coverUrl: string;
};

// Homepage
export type BlockType =
  | "hero"
  | "category_row"
  | "category_grid"
  | "bestsellers"
  | "deals"
  | "new_releases"
  | "gift_cards";

export type LayoutEntry = { id: string; type: BlockType; title: string };

export type Banner = {
  title: string;
  subtitle: string;
  ctaLabel: string;
  href: string;
  theme: "purple" | "yellow" | "mint" | "pink";
};
export type HeroData = { main: Banner; tiles: Banner[] };
export type Category = { slug: string; label: string; icon: string; href: string; count: number };
export type CategoryRowData = { categories: Category[] };
export type RailData = { products: ProductCard[]; seeAllHref: string };

export type Block<D> = LayoutEntry & { data: D };

// Cart
export type CartProblem = "OUT_OF_STOCK" | "NOT_ENOUGH_STOCK" | "OFFER_GONE";

export type CartLine = {
  offerId: string;
  productId: string;
  slug: string;
  name: string;
  platform: string;
  region: string;
  coverUrl: string;
  seller: { id: string; name: string };
  unitPriceCents: number;
  quantity: number;
  lineTotalCents: number;
  cashbackCents: number;
  available: number;
  problem?: CartProblem;
};

export type Cart = { items: CartLine[]; itemCount: number; totalCents: number; cashbackCents: number };

export type CartItemInput = { offerId: string; quantity: number };

export type MergeResult = {
  cart: Cart;
  dropped: { offerId: string; reason: "OFFER_GONE" | "OUT_OF_STOCK" | "CART_FULL" }[];
};

// Orders
export type OrderStatus = "pending_payment" | "paid" | "failed" | "expired";

export type OrderItem = {
  offerId: string;
  productId: string;
  name: string;
  slug: string;
  platform: string;
  region: string;
  sellerName: string;
  unitPriceCents: number;
  quantity: number;
  keys?: { code: string }[];
};

export type Order = {
  id: string;
  number: string;
  status: OrderStatus;
  items: OrderItem[];
  totalCents: number;
  cashbackCents: number;
  createdAt: string;
  reservedUntil: string;
  paidAt?: string;
  payment?: { last4: string; outcome: "success" | "declined" };
};

export type PayInput = { cardNumber: string; expiry: string; cvc: string; name: string };

// Wishlist
export type WishlistEntry = { product: ProductCard; addedAt: string };
export type Wishlist = { items: WishlistEntry[] };
export type WishlistMergeInput = { productId: string; addedAt: string };
