<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { login, fetchAuthConfig, user } from '../lib/auth.js';

  let email = '';
  let password = '';
  let error = null;
  let loading = false;
  let oidcEnabled = false;

  onMount(async () => {
    // If already logged in, go straight to home.
    const u = $user;
    if (u) { push('/'); return; }
    const cfg = await fetchAuthConfig();
    oidcEnabled = cfg.oidc_enabled;
  });

  async function handleSubmit(e) {
    e.preventDefault();
    error = null;
    loading = true;
    try {
      await login(email, password);
      push('/');
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-background px-4">
  <div class="w-full max-w-sm">
    <!-- Logo -->
    <div class="flex items-center justify-center gap-2.5 mb-8">
      <svg width="28" height="28" viewBox="0 0 32 32" xmlns="http://www.w3.org/2000/svg">
        <polygon points="16,2 27,8 27,22 16,28 5,22 5,8" fill="#E85D3A"/>
        <polygon points="16,9 22,13 22,21 16,25 10,21 10,13" fill="none" stroke="white" stroke-width="0.8" stroke-linejoin="round" opacity="0.4"/>
        <polygon points="16,14 19,16 19,19 16,21 13,19 13,16" fill="white" opacity="0.15"/>
      </svg>
      <span class="text-[20px] font-semibold tracking-tight">
        <span style="color:var(--text-muted)">Archi</span><span style="color:var(--foreground)">Pulse</span>
      </span>
    </div>

    <div class="bg-card border border-border rounded-xl p-6 shadow-sm">
      <h1 class="text-[16px] font-semibold text-foreground mb-1">Sign in</h1>
      <p class="text-[13px] text-muted-foreground mb-5">to continue to ArchiPulse</p>

      <form onsubmit={handleSubmit} class="space-y-4">
        <div>
          <label for="email" class="block text-[12px] font-medium text-foreground mb-1.5">Email</label>
          <input
            id="email"
            type="email"
            bind:value={email}
            required
            autocomplete="email"
            class="w-full bg-background border border-border rounded-md px-3 py-2 text-[13px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
            placeholder="you@example.com"
          />
        </div>

        <div>
          <label for="password" class="block text-[12px] font-medium text-foreground mb-1.5">Password</label>
          <input
            id="password"
            type="password"
            bind:value={password}
            required
            autocomplete="current-password"
            class="w-full bg-background border border-border rounded-md px-3 py-2 text-[13px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
            placeholder="••••••••"
          />
        </div>

        {#if error}
          <div class="text-[12px] text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">
            {error}
          </div>
        {/if}

        <button
          type="submit"
          disabled={loading}
          class="w-full bg-primary text-primary-foreground rounded-md py-2 text-[13px] font-medium hover:bg-primary/90 transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
        >
          {loading ? 'Signing in…' : 'Sign in'}
        </button>
      </form>

      {#if oidcEnabled}
        <div class="relative my-4">
          <div class="absolute inset-0 flex items-center"><div class="w-full border-t border-border"></div></div>
          <div class="relative flex justify-center text-[11px]"><span class="bg-card px-2 text-muted-foreground">or</span></div>
        </div>
        <a
          href="/api/v1/auth/oidc"
          class="flex items-center justify-center gap-2 w-full border border-border rounded-md py-2 text-[13px] text-foreground hover:bg-muted transition-colors"
        >
          Sign in with SSO
        </a>
      {/if}
    </div>
  </div>
</div>
