import type {
  UserInfo,
  LoginPayload,
  SignUpPayload,
  HttpErrorResponse,
  OAuthProvider,
} from "../types";

const BASE_URL = "https://api.bogdanantonovich.com/vox";

class ApiError extends Error {
  constructor(
    public code: number,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    credentials: "include", // sends cookies automatically (access_token, refresh_token)
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    ...options,
  });

  if (res.status === 204 || res.status === 201) {
    return undefined as T;
  }

  const data = await res.json().catch(() => null);

  if (!res.ok) {
    const errData = data as HttpErrorResponse | null;
    throw new ApiError(
      errData?.error?.code ?? res.status,
      errData?.error?.message ?? "Unknown error"
    );
  }

  return data as T;
}

// ─── Auth ───────────────────────────────────────────────

export const auth = {
  login: (payload: LoginPayload) =>
    request<Record<string, string>>("/auth/login", {
      method: "POST",
      body: JSON.stringify(payload),
    }),

  signUp: (payload: SignUpPayload) =>
    request<void>("/auth/sign_up", {
      method: "POST",
      body: JSON.stringify(payload),
    }),

  refresh: () =>
    request<void>("/auth/refresh", {
      method: "POST",
    }),

  /** Redirects browser to OAuth provider */
  oauthLogin: (provider: OAuthProvider) => {
    window.location.href = `${BASE_URL}/auth/${provider}/login`;
  },
};

// ─── User ────────────────────────────────────────────────

export const user = {
  getInfo: () => request<UserInfo>("/user/info"),

  uploadVoice: (audioBlob: Blob, textRef: string) =>
    request<void>(`/user/voice/new?text_ref=${encodeURIComponent(textRef)}`, {
      method: "POST",
      headers: { "Content-Type": "application/octet-stream" },
      body: audioBlob,
    }),
};

// ─── Hub ─────────────────────────────────────────────────

export const hub = {
  create: (hubId: string) =>
    request<Record<string, string>>(`/hub/${hubId}/new`, {
      method: "POST",
    }),

  /** Returns the full URL for listening — used with Audio/EventSource */
  listenUrl: (hubId: string) => `${BASE_URL}/hub/${hubId}/listen`,

  publish: (hubId: string, audioBlob: Blob, lang: string) =>
    request<void>(`/hub/${hubId}/publish?lang=${lang}`, {
      method: "POST",
      headers: { "Content-Type": "application/octet-stream" },
      body: audioBlob,
    }),
};

export { ApiError };
