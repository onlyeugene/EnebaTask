"use client";

import { useId, useRef, useState, type FormEvent, type KeyboardEvent } from "react";
import { useRouter } from "next/navigation";
import { Search, X } from "lucide-react";
import { MIN_CHARS, useAutocomplete } from "../hooks/use-autocomplete";
import { AutocompleteDropdown } from "./autocomplete-dropdown";

// Header search with suggestions as you type, like Eneba's. Enter opens
// /store?q=... (typo-tolerant search); a picked row opens its product page.
// Still a real GET form to /store, so Enter works before JS has loaded.
// `initialText` is the URL's q (see search-box.tsx).
export function SearchField({ initialText }: { initialText: string }) {
  const router = useRouter();
  const [text, setText] = useState(initialText);
  const [open, setOpen] = useState(false);
  // Highlighted row for arrow keys; -1 = none, the text box itself.
  const [active, setActive] = useState(-1);
  const inputRef = useRef<HTMLInputElement>(null);
  const listId = useId();

  const query = useAutocomplete(text);
  const suggestions = query.data ?? [];
  const typed = text.trim();
  const showList = open && typed.length >= MIN_CHARS;

  function go(href: string) {
    setOpen(false);
    setActive(-1);
    inputRef.current?.blur();
    router.push(href);
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (showList && active >= 0 && active < suggestions.length) {
      go(`/product/${suggestions[active].slug}`);
    } else {
      go(typed ? `/store?q=${encodeURIComponent(typed)}` : "/store");
    }
  }

  function onKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Escape") {
      setOpen(false);
      setActive(-1);
      return;
    }
    if (e.key !== "ArrowDown" && e.key !== "ArrowUp") return;
    if (!showList || suggestions.length === 0) return;
    e.preventDefault(); // keep the cursor where it is
    const n = suggestions.length;
    // Wraps: past the last row goes back to the text box, then the first.
    setActive((a) => (e.key === "ArrowDown" ? (a + 1 >= n ? -1 : a + 1) : a - 1 < -1 ? n - 1 : a - 1));
  }

  function clear() {
    setText("");
    setActive(-1);
    inputRef.current?.focus();
  }

  return (
    <>
      {/* Dims the page while the list is open. The form is lifted above it
          (z-50 over z-40), so only the box and its list stay bright. A click
          here moves focus off the input, and its blur closes the list. */}
      {showList && <div aria-hidden className="fixed inset-0 z-40 bg-brand-darkest/70" />}

      <form
        action="/store"
        role="search"
        onSubmit={onSubmit}
        className={`relative flex h-12 w-full items-center gap-3 border border-white px-2.5 md:h-[57px] ${showList ? "z-50 bg-brand" : ""}`}
      >
        <Search className="size-6 shrink-0 text-text-soft" aria-hidden />
        {/* type="text", not "search": the browser's own clear button would
            sit next to ours. */}
        <input
          ref={inputRef}
          name="q"
          type="text"
          enterKeyHint="search"
          maxLength={100}
          autoComplete="off"
          placeholder="Search for games, gift cards and more"
          aria-label="Search products"
          role="combobox"
          aria-expanded={showList}
          aria-controls={listId}
          aria-autocomplete="list"
          aria-activedescendant={showList && active >= 0 ? `${listId}-${active}` : undefined}
          value={text}
          onChange={(e) => {
            setText(e.target.value);
            setOpen(true);
            setActive(-1);
          }}
          onFocus={() => setOpen(true)}
          onBlur={() => setOpen(false)}
          onKeyDown={onKeyDown}
          className="h-full w-full bg-transparent text-base font-bold text-white outline-none placeholder:font-semibold placeholder:text-text-muted"
        />
        {text && (
          <button
            type="button"
            aria-label="Clear search"
            // Keeps focus in the input, so the list doesn't close and reopen.
            onMouseDown={(e) => e.preventDefault()}
            onClick={clear}
            className="link-hover shrink-0 p-1"
          >
            <X className="size-6" aria-hidden />
          </button>
        )}
        {showList && (
          <AutocompleteDropdown
            id={listId}
            text={typed}
            query={query}
            active={active}
            onHover={setActive}
            onPick={go}
          />
        )}
      </form>
    </>
  );
}
