import { useEffect, useState } from "react";

// Returns `value` once it has stopped changing for `ms` milliseconds.
export function useDebounce<T>(value: T, ms: number): T {
  const [settled, setSettled] = useState(value);
  useEffect(() => {
    const timer = setTimeout(() => setSettled(value), ms);
    return () => clearTimeout(timer);
  }, [value, ms]);
  return settled;
}
