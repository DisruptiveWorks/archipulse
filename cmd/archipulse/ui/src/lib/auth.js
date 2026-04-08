import { writable } from 'svelte/store';
import { push } from 'svelte-spa-router';

// Current authenticated user: { id, email, role } or null when logged out.
export const user = writable(null);

const BASE = '/api/v1';

// Fetch the current session from the server and populate the store.
// Resolves to the user object or null (never throws).
export async function fetchMe() {
  try {
    const r = await fetch(BASE + '/auth/me', { credentials: 'same-origin' });
    if (!r.ok) { user.set(null); return null; }
    const data = await r.json();
    user.set(data);
    return data;
  } catch {
    user.set(null);
    return null;
  }
}

// Log in with email + password. Throws on failure.
export async function login(email, password) {
  const r = await fetch(BASE + '/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'same-origin',
    body: JSON.stringify({ email, password }),
  });
  const data = await r.json();
  if (!r.ok) throw new Error(data.error || 'Login failed');
  user.set(data);
  return data;
}

// Log out and clear the store.
export async function logout() {
  await fetch(BASE + '/auth/logout', { method: 'POST', credentials: 'same-origin' });
  user.set(null);
  push('/login');
}

// Fetch OIDC config (whether OIDC is enabled).
export async function fetchAuthConfig() {
  try {
    const r = await fetch(BASE + '/auth/config');
    if (!r.ok) return { oidc_enabled: false };
    return r.json();
  } catch {
    return { oidc_enabled: false };
  }
}
