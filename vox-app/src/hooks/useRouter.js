import { useState, useEffect } from "react";

export function useRouter() {
  const [path, setPath] = useState(window.location.hash || "#/");
  useEffect(() => {
    const handler = () => setPath(window.location.hash || "#/");
    window.addEventListener("hashchange", handler);
    return () => window.removeEventListener("hashchange", handler);
  }, []);
  const navigate = (to) => {
    window.location.hash = to;
  };
  return { path, navigate };
}
