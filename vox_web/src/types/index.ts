// API Types derived from Swagger spec

export interface UserInfo {
  id: string;
  email: string;
  name: string;
  picture: string;
}

export interface LoginPayload {
  login: string;
  password: string;
}

export interface SignUpPayload {
  email: string;
  login: string;
  name: string;
  password: string;
}

export interface HttpErrorResponse {
  error: {
    code: number;
    message: string;
  };
}

export type OAuthProvider = "google" | "github";

export interface Language {
  code: string;
  label: string;
}

// App state types
export interface AuthState {
  user: UserInfo | null;
  isLoading: boolean;
}

export type Route =
  | "#/"
  | "#/login"
  | "#/signup"
  | "#/admin"
  | "#/room"
  | string; // for dynamic routes like #/room/abc
