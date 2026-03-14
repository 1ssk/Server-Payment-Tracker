import { VPNServer, SMTPSettings } from '../types/server';

const API_BASE = '';

async function request<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
  const res = await fetch(input, {
    headers: {
      'Content-Type': 'application/json',
      ...(init && init.headers),
    },
    ...init,
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || `Request failed with status ${res.status}`);
  }

  return res.json() as Promise<T>;
}

export async function apiLogin(username: string, password: string): Promise<{ token: string }> {
  return request<{ token: string }>(`${API_BASE}/api/login`, {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export async function apiGetServers(): Promise<VPNServer[]> {
  const data = await request<{ servers: VPNServer[] }>(`${API_BASE}/api/servers`, {
    method: 'GET',
  });
  return data.servers;
}

export async function apiCreateServer(server: Omit<VPNServer, 'id' | 'createdAt'>): Promise<VPNServer> {
  return request<VPNServer>(`${API_BASE}/api/servers`, {
    method: 'POST',
    body: JSON.stringify(server),
  });
}

export async function apiUpdateServer(server: VPNServer): Promise<VPNServer> {
  return request<VPNServer>(`${API_BASE}/api/servers`, {
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
  return request<SMTPSettings>(`${API_BASE}/api/smtp-settings`, {
    method: 'PUT',
    body: JSON.stringify(settings),
  });
}

