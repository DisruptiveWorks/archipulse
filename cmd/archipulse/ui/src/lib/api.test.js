import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock svelte-spa-router before importing api
vi.mock('svelte-spa-router', () => ({ push: vi.fn() }));

import { api } from './api.js';
import { push } from 'svelte-spa-router';

function mockFetch(status, body) {
  return vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
    status,
    ok: status >= 200 && status < 300,
    statusText: status === 200 ? 'OK' : 'Error',
    json: async () => body,
  });
}

beforeEach(() => {
  vi.restoreAllMocks();
});

describe('api.get', () => {
  it('returns parsed JSON on success', async () => {
    mockFetch(200, { id: 1, name: 'test' });
    const result = await api.get('/workspaces');
    expect(result).toEqual({ id: 1, name: 'test' });
  });

  it('redirects to /login and throws on 401', async () => {
    mockFetch(401, { error: 'unauthorized' });
    await expect(api.get('/workspaces')).rejects.toThrow('Not authenticated');
    expect(push).toHaveBeenCalledWith('/login');
  });

  it('throws error message from response body on non-ok', async () => {
    mockFetch(404, { error: 'not found' });
    await expect(api.get('/workspaces/bad')).rejects.toThrow('not found');
  });

  it('sends credentials: same-origin', async () => {
    const spy = mockFetch(200, {});
    await api.get('/workspaces');
    expect(spy).toHaveBeenCalledWith('/api/v1/workspaces', { credentials: 'same-origin' });
  });
});

describe('api.post', () => {
  it('sends JSON body and returns parsed response', async () => {
    const spy = mockFetch(200, { id: 42 });
    const result = await api.post('/workspaces', { name: 'New WS' });
    expect(result).toEqual({ id: 42 });
    expect(spy).toHaveBeenCalledWith('/api/v1/workspaces', expect.objectContaining({
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: 'New WS' }),
    }));
  });
});

describe('api.put', () => {
  it('sends PUT with JSON body', async () => {
    const spy = mockFetch(200, { updated: true });
    await api.put('/workspaces/1', { name: 'Updated' });
    expect(spy).toHaveBeenCalledWith('/api/v1/workspaces/1', expect.objectContaining({
      method: 'PUT',
      body: JSON.stringify({ name: 'Updated' }),
    }));
  });
});

describe('api.upload', () => {
  it('sends POST with FormData (no Content-Type header)', async () => {
    const spy = mockFetch(200, { ok: true });
    const fd = new FormData();
    await api.upload('/workspaces/1/import', fd);
    const call = spy.mock.calls[0][1];
    expect(call.method).toBe('POST');
    expect(call.body).toBe(fd);
    expect(call.headers).toBeUndefined();
  });
});
