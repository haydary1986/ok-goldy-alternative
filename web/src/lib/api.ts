// Tiny fetch wrapper that handles the standard Goldy API envelope and
// attaches the X-Goldy-Actor header so server-side audit rows are not all
// "system". The actor email lives in localStorage until proper OIDC login
// is wired in.

const API_BASE = '/api/v1';

interface Envelope<T> {
  success: boolean;
  data?: T;
  error?: { code: string; message: string };
  meta?: {
    total?: number;
    page?: number;
    page_size?: number;
    next_page_token?: string;
  };
}

const ACTOR_KEY = 'goldy_actor';

export function getActor(): string {
  return localStorage.getItem(ACTOR_KEY) ?? '';
}

export function setActor(email: string): void {
  if (email) {
    localStorage.setItem(ACTOR_KEY, email);
  } else {
    localStorage.removeItem(ACTOR_KEY);
  }
}

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  headers.set('Accept', 'application/json');
  if (init?.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const actor = getActor();
  if (actor) {
    headers.set('X-Goldy-Actor', actor);
  }

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });

  let body: Envelope<T>;
  try {
    body = (await res.json()) as Envelope<T>;
  } catch {
    throw new Error(`HTTP ${res.status} (no JSON body)`);
  }

  if (!body.success || body.error) {
    throw new Error(body.error?.message ?? `HTTP ${res.status}`);
  }
  return body.data as T;
}
