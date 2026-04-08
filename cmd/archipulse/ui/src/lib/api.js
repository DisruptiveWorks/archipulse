import { push } from 'svelte-spa-router';

const BASE = '/api/v1';

async function handleResponse(r) {
  if (r.status === 401) {
    push('/login');
    throw new Error('Not authenticated');
  }
  if (!r.ok) {
    let msg = r.statusText;
    try { msg = (await r.json()).error || msg; } catch { /* ignore */ }
    throw new Error(msg);
  }
  return r.json();
}

export const api = {
  async get(path) {
    const r = await fetch(BASE + path, { credentials: 'same-origin' });
    return handleResponse(r);
  },
  async post(path, body) {
    const r = await fetch(BASE + path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'same-origin',
      body: JSON.stringify(body),
    });
    return handleResponse(r);
  },
  async put(path, body) {
    const r = await fetch(BASE + path, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'same-origin',
      body: JSON.stringify(body),
    });
    return handleResponse(r);
  },
  async upload(path, formData) {
    const r = await fetch(BASE + path, {
      method: 'POST',
      credentials: 'same-origin',
      body: formData,
    });
    return handleResponse(r);
  },
};
