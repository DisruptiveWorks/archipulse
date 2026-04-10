import { describe, it, expect, vi, beforeEach } from 'vitest';
import { get } from 'svelte/store';

vi.mock('svelte-spa-router', () => ({ push: vi.fn() }));

import { user, fetchMe, login, logout, fetchAuthConfig } from './auth.js';
import { push } from 'svelte-spa-router';

function mockFetch(status, body) {
  return vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
    status,
    ok: status >= 200 && status < 300,
    json: async () => body,
  });
}

beforeEach(() => {
  vi.restoreAllMocks();
  user.set(null);
});

describe('fetchMe', () => {
  it('sets user store and returns user on success', async () => {
    mockFetch(200, { id: '1', email: 'a@b.com', role: 'admin' });
    const result = await fetchMe();
    expect(result).toEqual({ id: '1', email: 'a@b.com', role: 'admin' });
    expect(get(user)).toEqual({ id: '1', email: 'a@b.com', role: 'admin' });
  });

  it('sets user to null and returns null on 401', async () => {
    mockFetch(401, {});
    const result = await fetchMe();
    expect(result).toBeNull();
    expect(get(user)).toBeNull();
  });

  it('returns null without throwing on fetch error', async () => {
    vi.spyOn(globalThis, 'fetch').mockRejectedValueOnce(new Error('network error'));
    const result = await fetchMe();
    expect(result).toBeNull();
  });
});

describe('login', () => {
  it('sets user store and returns data on success', async () => {
    mockFetch(200, { id: '1', email: 'a@b.com', role: 'viewer' });
    const result = await login('a@b.com', 'pass');
    expect(result.email).toBe('a@b.com');
    expect(get(user)?.email).toBe('a@b.com');
  });

  it('throws error message on failed login', async () => {
    mockFetch(401, { error: 'invalid credentials' });
    await expect(login('bad@b.com', 'wrong')).rejects.toThrow('invalid credentials');
    expect(get(user)).toBeNull();
  });
});

describe('logout', () => {
  it('clears user store and redirects to /login', async () => {
    user.set({ id: '1', email: 'a@b.com', role: 'admin' });
    mockFetch(200, {});
    await logout();
    expect(get(user)).toBeNull();
    expect(push).toHaveBeenCalledWith('/login');
  });
});

describe('fetchAuthConfig', () => {
  it('returns config with oidc_enabled true', async () => {
    mockFetch(200, { oidc_enabled: true, demo_mode: false });
    const cfg = await fetchAuthConfig();
    expect(cfg.oidc_enabled).toBe(true);
  });

  it('returns oidc_enabled: false on error', async () => {
    vi.spyOn(globalThis, 'fetch').mockRejectedValueOnce(new Error('network'));
    const cfg = await fetchAuthConfig();
    expect(cfg.oidc_enabled).toBe(false);
  });

  it('returns oidc_enabled: false on non-ok response', async () => {
    mockFetch(500, {});
    const cfg = await fetchAuthConfig();
    expect(cfg.oidc_enabled).toBe(false);
  });
});
