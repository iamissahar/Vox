import axios from 'axios';
import type {
  LoginPayload,
  SignUpPayload,
  UserInfo,
  VoiceReference,
} from '../types';

// ─── Client ──────────────────────────────────────────────────────────────────

export const apiClient = axios.create({
  baseURL: 'https://bogdanantonovich.com/vox/api',
  withCredentials: true,
});

// ─── Auth ────────────────────────────────────────────────────────────────────

export const authApi = {
  login: (payload: LoginPayload) =>
    apiClient.post('/auth/login', payload),

  signUp: (payload: SignUpPayload) =>
    apiClient.post('/auth/sign_up', payload),

  refresh: () =>
    apiClient.post('/auth/refresh'),

  oauthLogin: (provider: 'google' | 'github') => {
    window.location.href = `https://bogdanantonovich.com/vox/api/auth/${provider}/login`;
  },
};

// ─── User ────────────────────────────────────────────────────────────────────

export const userApi = {
  getInfo: () =>
    apiClient.get<UserInfo>('/user/info'),

  getHubs: () =>
    apiClient.get<{ hub_ids: string[] }>('/user/hubs'),
};

// ─── Hub ─────────────────────────────────────────────────────────────────────

export const hubApi = {
  create: () =>
    apiClient.post<{ hub_id: string }>('/hub'),

  delete: (hubId: string, userId: string) =>
    apiClient.delete(`/hub/${hubId}`, { data: { user_id: userId } }),

  reconnect: (hubId: string) =>
    apiClient.get(`/hub/${hubId}/reconnect`),

  getListenUrl: (hubId: string) =>
    `https://bogdanantonovich.com/vox/api/hub/${hubId}/listen`,

  getPublishUrl: (hubId: string, lang: string = 'ru') =>
    `https://bogdanantonovich.com/vox/api/hub/${hubId}/publish?lang=${lang}`,
};

// ─── Voice ───────────────────────────────────────────────────────────────────

export const voiceApi = {
  getMeta: () =>
    apiClient.get<VoiceReference[]>('/user/voice/meta'),

  getFile: (fileId: string) =>
    apiClient.get('/user/voice/file', {
      params: { file_id: fileId },
      responseType: 'blob',
    }),

  upload: (audioBlob: Blob, textRef: string) => {
    return apiClient.post('/user/voice', audioBlob, {
      headers: { 'Content-Type': 'application/octet-stream' },
      params: { text_ref: textRef },
    });
  },

  delete: (fileId: string) =>
    apiClient.delete('/user/voice', { params: { file_id: fileId } }),
};
