// ─── Auth ────────────────────────────────────────────────────────────────────

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

// ─── User ────────────────────────────────────────────────────────────────────

export interface UserInfo {
  id: string;
  email: string;
  name: string;
  picture: string;
}

// ─── Hub ─────────────────────────────────────────────────────────────────────

export interface Hub {
  id: string;
}

// ─── Voice ───────────────────────────────────────────────────────────────────

export interface VoiceReference {
  file_id: string;
  path: string;
  text: string;
  type: string;
}

// ─── HTTP ────────────────────────────────────────────────────────────────────

export interface HttpErrorResponse {
  error: {
    code: number;
    message: string;
  };
}
