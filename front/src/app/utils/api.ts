import { Server, SMTPSettings, Payment, ReportRow } from '../types/server';

const API_BASE = '';
const AUTH_KEY = 'app-auth-token';

async function request<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
  const token = typeof window !== 'undefined' ? localStorage.getItem(AUTH_KEY) : null;

  const res = await fetch(input, {
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(init && init.headers),
    },
    ...init,
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || `Request failed with status ${res.status}`);
  }

  // Для 204 No Content возвращаем undefined.
  if (res.status === 204) {
    return undefined as T;
  }

  return res.json() as Promise<T>;
}

export async function apiLogin(username: string, password: string): Promise<{ token: string }> {
  return request<{ token: string }>(`${API_BASE}/api/login`, {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export async function apiGetServers(): Promise<Server[]> {
  const data = await request<{ servers: Server[] }>(`${API_BASE}/api/servers`, {
    method: 'GET',
  });
  return data.servers ?? [];
}

export async function apiCreateServer(server: Omit<Server, 'id' | 'createdAt'>): Promise<Server> {
  return request<Server>(`${API_BASE}/api/servers`, {
    method: 'POST',
    body: JSON.stringify(server),
  });
}

export async function apiUpdateServer(server: Server): Promise<Server> {
  return request<Server>(`${API_BASE}/api/servers`, {
    method: 'PUT',
    body: JSON.stringify(server),
  });
}

export async function apiDeleteServer(id: string): Promise<void> {
  await request<void>(`${API_BASE}/api/servers?id=${encodeURIComponent(id)}`, {
    method: 'DELETE',
  });
}

export async function apiGetSMTPSettings(): Promise<SMTPSettings> {
  return request<SMTPSettings>(`${API_BASE}/api/smtp-settings`, {
    method: 'GET',
  });
}

export async function apiUpdateSMTPSettings(settings: SMTPSettings): Promise<SMTPSettings> {
  const payload = { ...settings };
  if (payload.reminderDaysBefore == null || payload.reminderDaysBefore < 1) {
    payload.reminderDaysBefore = 10;
  }
  return request<SMTPSettings>(`${API_BASE}/api/smtp-settings`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export async function apiConfirmPayment(serverId: string, paidAt: string, amount?: number): Promise<Payment> {
  return request<Payment>(`${API_BASE}/api/servers/${encodeURIComponent(serverId)}/payments`, {
    method: 'POST',
    body: JSON.stringify({ paidAt, amount }),
  });
}

export async function apiGetServerPayments(serverId: string, from?: string, to?: string): Promise<Payment[]> {
  const params = new URLSearchParams();
  if (from) params.set('from', from);
  if (to) params.set('to', to);
  const q = params.toString();
  const data = await request<{ payments: Payment[] }>(
    `${API_BASE}/api/servers/${encodeURIComponent(serverId)}/payments${q ? `?${q}` : ''}`,
    { method: 'GET' }
  );
  return data.payments ?? [];
}

export async function apiGetReports(params: { from?: string; to?: string; serverId?: string }): Promise<ReportRow[]> {
  const search = new URLSearchParams();
  if (params.from) search.set('from', params.from);
  if (params.to) search.set('to', params.to);
  if (params.serverId) search.set('serverId', params.serverId);
  const q = search.toString();
  const data = await request<{ rows: ReportRow[] }>(`${API_BASE}/api/reports${q ? `?${q}` : ''}`, { method: 'GET' });
  return data.rows ?? [];
}

export async function apiSendTestEmail(): Promise<void> {
  await request<{ ok: boolean }>(`${API_BASE}/api/smtp-settings/test`, { method: 'POST' });
}

