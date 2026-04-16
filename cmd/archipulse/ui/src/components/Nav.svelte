<script>
  import { push } from 'svelte-spa-router';
  import { Button } from '$lib/components/ui/button';
  import { createEventDispatcher } from 'svelte';
  import { user, logout } from '../lib/auth.js';

  export let wsId = null;
  export let wsName = null;
  export let viewLabel = null;

  const dispatch = createEventDispatcher();

  let userMenuOpen = false;

  function showCreateWs() {
    window.dispatchEvent(new CustomEvent('archipulse:create-ws'));
  }

  function toggleSidebar() {
    dispatch('toggleSidebar');
  }

  function toggleUserMenu() {
    userMenuOpen = !userMenuOpen;
  }

  function closeUserMenu() {
    userMenuOpen = false;
  }

  async function handleLogout() {
    userMenuOpen = false;
    await logout();
  }
</script>

<svelte:window onclick={(e) => {
  if (!e.target.closest('.user-menu-wrap')) userMenuOpen = false;
}} />

<nav>
  {#if wsId}
    <button class="nav-hamburger" onclick={toggleSidebar} aria-label="Toggle menu">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
        <line x1="3" y1="6" x2="21" y2="6"/>
        <line x1="3" y1="12" x2="21" y2="12"/>
        <line x1="3" y1="18" x2="21" y2="18"/>
      </svg>
    </button>
  {/if}
  <a class="nav-logo" href="#/">
    <svg width="22" height="22" viewBox="0 0 32 32" xmlns="http://www.w3.org/2000/svg">
      <polygon points="16,2 27,8 27,22 16,28 5,22 5,8" fill="#E85D3A"/>
      <polygon points="16,9 22,13 22,21 16,25 10,21 10,13" fill="none" stroke="white" stroke-width="0.8" stroke-linejoin="round" opacity="0.4"/>
      <polygon points="16,14 19,16 19,19 16,21 13,19 13,16" fill="white" opacity="0.15"/>
    </svg>
    <div class="nav-wordmark"><span class="w-archi">Archi</span><span class="w-pulse">Pulse</span></div>
  </a>
  <div class="nav-sep"></div>
  <div class="breadcrumb">
    {#if !wsId}
      <span style="color:var(--text-muted)">Workspaces</span>
    {:else if !viewLabel}
      <a href="#/">Workspaces</a>
      <span class="sep">/</span>
      <span class="current">{wsName || wsId}</span>
    {:else}
      <a href="#/">Workspaces</a>
      <span class="sep">/</span>
      <a href="#/ws/{wsId}">{wsName || wsId}</a>
      <span class="sep">/</span>
      <span class="current">{viewLabel}</span>
    {/if}
  </div>
  <div class="nav-spacer"></div>

  <!-- User menu -->
  {#if $user}
    {#if !wsId && $user.org_role === 'admin'}
      <Button size="sm" onclick={showCreateWs}>+ New workspace</Button>
    {/if}
    <div class="user-menu-wrap" style="position:relative;">
      <button
        class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-[13px] text-foreground hover:bg-muted transition-colors border border-border bg-card"
        onclick={toggleUserMenu}
        aria-label="User menu"
      >
        <span class="size-5 rounded-full bg-primary/20 text-primary flex items-center justify-center text-[11px] font-bold flex-shrink-0">
          {$user.email[0].toUpperCase()}
        </span>
        <span class="hidden sm:inline max-w-[120px] truncate">{$user.email}</span>
        <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.8" class="flex-shrink-0 opacity-60">
          <path d="M2 4l4 4 4-4"/>
        </svg>
      </button>

      {#if userMenuOpen}
        <div class="absolute right-0 top-full mt-1 w-52 bg-popover border border-border rounded-lg shadow-lg py-1 z-50 text-[13px]">
          <div class="px-3 py-2 border-b border-border">
            <div class="font-medium text-foreground truncate">{$user.email}</div>
            <div class="text-[11px] text-muted-foreground capitalize">{$user.org_role}</div>
          </div>
          <button
            class="w-full text-left px-3 py-2 text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
            onclick={handleLogout}
          >Sign out</button>
        </div>
      {/if}
    </div>
  {/if}
</nav>
