import { useState, useEffect } from "react";
import { user as userApi } from "../api/client";
import type { UserInfo } from "../types";

interface UseAuthReturn {
  user: UserInfo | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  setUser: (u: UserInfo | null) => void;
  logout: () => void;
}

export function useAuth(): UseAuthReturn {
  const [currentUser, setCurrentUser] = useState<UserInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    userApi
      .getInfo()
      .then((info) => setCurrentUser(info))
      .catch(() => setCurrentUser(null))
      .finally(() => setIsLoading(false));
  }, []);

  const logout = () => {
    // Cookies are cleared server-side; just reset state
    setCurrentUser(null);
  };

  return {
    user: currentUser,
    isLoading,
    isAuthenticated: !!currentUser,
    setUser: setCurrentUser,
    logout,
  };
}
