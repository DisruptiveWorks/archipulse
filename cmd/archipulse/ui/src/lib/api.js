const BASE = '/api/v1';

export const api = {
  async get(path) {
    const r = await fetch(BASE + path);
    if (!r.ok) throw new Error((await r.json()).error || r.statusText);
    return r.json();
  },
  async post(path, body) {
    const r = await fetch(BASE + path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (!r.ok) throw new Error((await r.json()).error || r.statusText);
    return r.json();
  },
  async put(path, body) {
    const r = await fetch(BASE + path, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (!r.ok) throw new Error((await r.json()).error || r.statusText);
    return r.json();
  },
  async upload(path, formData) {
    const r = await fetch(BASE + path, { method: 'POST', body: formData });
    if (!r.ok) throw new Error((await r.json()).error || r.statusText);
    return r.json();
  },
};
