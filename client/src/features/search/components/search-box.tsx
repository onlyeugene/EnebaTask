"use client";

import { Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { SearchField } from "./search-field";

// The header search. It shows the URL's q, so /store?q=... opened directly
// (a link, a reload) has its text in the box.
// Next needs a Suspense boundary around anything that reads the URL on a
// static page. The fallback is the same box, empty, so the first HTML
// looks the same and Enter still works before JS loads.
export function SearchBox() {
  return (
    <Suspense fallback={<SearchField initialText="" />}>
      <UrlSearchField />
    </Suspense>
  );
}

function UrlSearchField() {
  const q = useSearchParams().get("q") ?? "";
  // The key starts a fresh box whenever q changes (a new search, Back, or
  // leaving the store), so the text always matches the URL. Typing doesn't
  // change q, so it never resets text mid-typing.
  return <SearchField key={q} initialText={q} />;
}
