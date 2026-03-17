import { useState, useEffect, useCallback } from "react";

export function useRouter() {
  const [route, setRoute] = useState<string>(
    window.location.hash || "#/"
  );

  useEffect(() => {
    const handler = () => setRoute(window.location.hash || "#/");
    window.addEventListener("hashchange", handler);
    return () => window.removeEventListener("hashchange", handler);
  }, []);

  const navigate = useCallback((to: string) => {
    window.location.hash = to;
  }, []);

  return { route, navigate };
}
