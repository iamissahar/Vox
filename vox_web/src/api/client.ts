import type {
  UserInfo,
  LoginPayload,
  SignUpPayload,
  HttpErrorResponse,
  OAuthProvider,
} from "../types";

const BASE_URL = "https://bogdanantonovich.com/vox/api";

class ApiError extends Error {
  constructor(
    public code: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

const MISSING_COOKIE_CODE = 3;

async function request<T>(
  path: string,
  options: RequestInit = {},
  isRetry = false,
): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    ...options,
  });

  if (res.status === 204) {
    return undefined as T;
  }

  const data = await res.json().catch(() => null);

  if (!res.ok) {
    const errData = data as HttpErrorResponse | null;
    const code = errData?.error?.code ?? res.status;
    const message = errData?.error?.message ?? "Unknown error";

    if (code === MISSING_COOKIE_CODE && !isRetry) {
      await request<void>("/auth/refresh", { method: "POST" });
      return request<T>(path, options, true);
    }

    throw new ApiError(code, message);
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
  // POST /hub — user из куки, body не нужен
  createAuto: () =>
    request<{ hub_id: string }>("/hub", {
      method: "POST",
    }),

  // GET /user/hubs — user из куки, body не нужен
  listMine: () => request<{ hub_ids: string[] }>("/user/hubs"),

  // DELETE /hub/{hub_id} — body: { user_id }
  delete: (hubId: string, userId: string) =>
    request<void>(`/hub/${hubId}`, {
      method: "DELETE",
      body: JSON.stringify({ user_id: userId }),
    }),

  // GET /hub/{hub_id}/reconnect — user из куки, редирект на фронтенд
  reconnect: (hubId: string) => {
    window.location.href = `${BASE_URL}/hub/${hubId}/reconnect`;
  },

  listenUrl: (hubId: string) => `${BASE_URL}/hub/${hubId}/listen`,

  publish: (hubId: string, audioBlob: Blob, lang: string) =>
    request<void>(`/hub/${hubId}/publish?lang=${lang}`, {
      method: "POST",
      headers: { "Content-Type": "application/octet-stream" },
      body: audioBlob,
    }),
};

export { ApiError };
