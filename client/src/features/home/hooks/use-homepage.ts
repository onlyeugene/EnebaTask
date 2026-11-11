import { useQuery } from "@tanstack/react-query";
import { getHomepageLayout } from "../api";

export function useHomepage() {
  return useQuery({ queryKey: ["homepage"], queryFn: getHomepageLayout });
}
