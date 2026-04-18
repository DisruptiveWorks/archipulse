<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { api } from '../lib/api.js';
  import { user } from '../lib/auth.js';

  export let params = {};

  $: wsId = params.wsId;

  // ── Tabs ──────────────────────────────────────────────────────────────────
  let tab = 'members'; // 'general' | 'members'

  // ── Workspace general ─────────────────────────────────────────────────────
  let ws = null;
  let wsLoading = true;
  let wsSaving = false;
  let wsError = null;
  let wsSuccess = false;
  let wsForm = { name: '', purpose: '', description: '' };

  const PURPOSES = ['as-is', 'to-be', 'initiative', 'other'];

  // ── Members ───────────────────────────────────────────────────────────────
  let members = [];
  let membersLoading = true;
  let membersError = null;

  // Add member form
  let addEmail = '';
  let addRole = 'viewer';
  let addSearching = false;
  let addError = null;
  let addSuccess = false;

  // Inline role updating
  let updatingRole = {}; // userId → true while in flight

  // Removing
  let removingMember = {}; // userId → true while in flight

  // ── Load ──────────────────────────────────────────────────────────────────
  onMount(async () => {
    await Promise.all([loadWs(), loadMembers()]);
  });

  async function loadWs() {
    wsLoading = true;
    wsError = null;
    try {
      ws = await api.get('/workspaces/' + wsId);
      wsForm = { name: ws.name, purpose: ws.purpose, description: ws.description || '' };
    } catch (e) {
      wsError = e.message;
    }
    wsLoading = false;
  }

  async function loadMembers() {
    membersLoading = true;
    membersError = null;
    try {
      members = await api.get('/workspaces/' + wsId + '/members');
    } catch (e) {
      membersError = e.message;
    }
    membersLoading = false;
  }

  // ── General form ──────────────────────────────────────────────────────────
  async function saveGeneral() {
    wsSaving = true;
    wsError = null;
    wsSuccess = false;
    try {
      ws = await api.put('/workspaces/' + wsId, { ...wsForm, version: ws.version });
      wsForm = { name: ws.name, purpose: ws.purpose, description: ws.description || '' };
      wsSuccess = true;
      setTimeout(() => wsSuccess = false, 2500);
    } catch (e) {
      wsError = e.message;
    }
    wsSaving = false;
  }

  // ── Add member ────────────────────────────────────────────────────────────
  async function addMember() {
    if (!addEmail.trim()) return;
    addSearching = true;
    addError = null;
    addSuccess = false;
    try {
      const found = await api.get('/users/lookup?email=' + encodeURIComponent(addEmail.trim()));
      await api.post('/workspaces/' + wsId + '/members', { user_id: found.id, role: addRole });
      addSuccess = true;
      addEmail = '';
      setTimeout(() => addSuccess = false, 2500);
      await loadMembers();
    } catch (e) {
      addError = e.message;
    }
    addSearching = false;
  }

  // ── Update role ───────────────────────────────────────────────────────────
  async function updateRole(member, newRole) {
    if (newRole === member.role) return;
    updatingRole[member.user_id] = true;
    try {
      await api.put('/workspaces/' + wsId + '/members/' + member.user_id, { role: newRole });
      member.role = newRole;
      members = [...members];
    } catch (e) {
      // silently revert — reload to sync
      await loadMembers();
    }
    updatingRole[member.user_id] = false;
  }

  // ── Remove member ─────────────────────────────────────────────────────────
  async function removeMember(member) {
    removingMember[member.user_id] = true;
    try {
      await api.delete('/workspaces/' + wsId + '/members/' + member.user_id);
      members = members.filter(m => m.user_id !== member.user_id);
    } catch (e) {
      membersError = e.message;
    }
    removingMember[member.user_id] = false;
  }

  // ── Helpers ───────────────────────────────────────────────────────────────
  const ROLES = ['viewer', 'editor', 'owner'];
  const ROLE_LABELS = { viewer: 'Viewer', editor: 'Editor', owner: 'Owner' };
  const ROLE_COLORS = {
    owner:  'bg-amber-100 text-amber-800 border-amber-200',
    editor: 'bg-blue-100 text-blue-800 border-blue-200',
    viewer: 'bg-gray-100 text-gray-600 border-gray-200',
  };

  $: currentUserId = $user?.id;
  $: myRole = members.find(m => m.user_id === currentUserId)?.role;
  $: canManage = $user?.org_role === 'admin' || myRole === 'owner';
  $: isOwner = myRole === 'owner';

  // ── Delete workspace ──────────────────────────────────────────────────────
  let deleteConfirm = '';
  let deleting = false;
  let deleteError = null;

  $: deleteReady = ws && deleteConfirm.trim() === ws.name.trim();

  async function deleteWorkspace() {
    if (!deleteReady || deleting) return;
    deleting = true;
    deleteError = null;
    try {
      await api.delete('/workspaces/' + wsId);
      push('/');
    } catch (e) {
      deleteError = e.message;
      deleting = false;
    }
  }
</script>

<div class="content max-w-2xl">
  <div class="mb-6">
    <h1 class="text-[18px] font-semibold">Workspace settings</h1>
    {#if ws}<p class="text-muted-foreground text-[13px] mt-0.5">{ws.name}</p>{/if}
  </div>

  <!-- Tabs -->
  <div class="flex gap-0 border-b border-border mb-6">
    {#each [['members', 'Members'], ['general', 'General']] as [key, label]}
      <button
        class="px-4 py-2 text-[13px] font-medium border-b-2 transition-colors -mb-px
          {tab === key ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'}"
        on:click={() => tab = key}
      >{label}</button>
    {/each}
  </div>

  <!-- ── Members tab ──────────────────────────────────────────────────── -->
  {#if tab === 'members'}
    {#if membersLoading}
      <div class="flex items-center gap-2 text-muted-foreground py-6">
        <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
        Loading…
      </div>
    {:else}
      {#if membersError}
        <div class="mb-4 text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">{membersError}</div>
      {/if}

      <!-- Member list -->
      <div class="bg-card border border-border rounded-lg overflow-hidden mb-6">
        {#if members.length === 0}
          <div class="px-4 py-6 text-center text-muted-foreground text-[13px]">No members yet.</div>
        {:else}
          {#each members as member (member.user_id)}
            <div class="flex items-center gap-3 px-4 py-3 border-b border-border last:border-0">
              <!-- Avatar initials -->
              <div class="size-8 rounded-full bg-muted flex items-center justify-center text-[12px] font-semibold text-muted-foreground flex-shrink-0 uppercase">
                {member.email[0]}
              </div>

              <!-- Email -->
              <div class="flex-1 min-w-0">
                <div class="text-[13px] font-medium truncate">{member.email}</div>
                {#if member.user_id === currentUserId}
                  <div class="text-[11px] text-muted-foreground">You</div>
                {/if}
              </div>

              <!-- Role selector -->
              {#if canManage && member.user_id !== currentUserId}
                <div class="relative">
                  {#if updatingRole[member.user_id]}
                    <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin"></div>
                  {:else}
                    <select
                      class="text-[12px] border border-border rounded-md px-2 py-1 bg-background cursor-pointer focus:outline-none focus:ring-1 focus:ring-primary"
                      value={member.role}
                      on:change={e => updateRole(member, e.target.value)}
                    >
                      {#each ROLES as r}
                        <option value={r}>{ROLE_LABELS[r]}</option>
                      {/each}
                    </select>
                  {/if}
                </div>
              {:else}
                <span class="text-[11px] font-semibold px-2 py-0.5 rounded-full border {ROLE_COLORS[member.role] || ''}">
                  {ROLE_LABELS[member.role] || member.role}
                </span>
              {/if}

              <!-- Remove button -->
              {#if canManage && member.user_id !== currentUserId}
                <button
                  class="text-muted-foreground hover:text-destructive transition-colors ml-1 disabled:opacity-40"
                  disabled={removingMember[member.user_id]}
                  on:click={() => removeMember(member)}
                  title="Remove member"
                >
                  {#if removingMember[member.user_id]}
                    <div class="size-3.5 rounded-full border-2 border-border border-t-destructive animate-spin"></div>
                  {:else}
                    <svg xmlns="http://www.w3.org/2000/svg" class="size-4" viewBox="0 0 20 20" fill="currentColor">
                      <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"/>
                    </svg>
                  {/if}
                </button>
              {/if}
            </div>
          {/each}
        {/if}
      </div>

      <!-- Add member -->
      {#if canManage}
        <div class="bg-card border border-border rounded-lg p-4">
          <div class="text-[12px] font-semibold uppercase tracking-wide text-muted-foreground mb-3">Add member</div>
          <div class="flex gap-2">
            <input
              class="flex-1 text-[13px] border border-border rounded-md px-3 py-1.5 bg-background focus:outline-none focus:ring-1 focus:ring-primary"
              type="email"
              placeholder="user@example.com"
              bind:value={addEmail}
              on:keydown={e => e.key === 'Enter' && addMember()}
            />
            <select
              class="text-[13px] border border-border rounded-md px-2 py-1.5 bg-background focus:outline-none focus:ring-1 focus:ring-primary"
              bind:value={addRole}
            >
              {#each ROLES as r}
                <option value={r}>{ROLE_LABELS[r]}</option>
              {/each}
            </select>
            <button
              class="px-3 py-1.5 bg-primary text-primary-foreground text-[13px] font-medium rounded-md hover:bg-primary/90 transition-colors disabled:opacity-50"
              disabled={addSearching || !addEmail.trim()}
              on:click={addMember}
            >
              {addSearching ? '…' : 'Add'}
            </button>
          </div>
          {#if addError}
            <div class="mt-2 text-[12px] text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-1.5">{addError}</div>
          {/if}
          {#if addSuccess}
            <div class="mt-2 text-[12px] text-green-700 bg-green-50 border border-green-200 rounded-md px-3 py-1.5">Member added.</div>
          {/if}
          <p class="text-[11px] text-muted-foreground mt-2">The user must have logged in at least once to appear in the system.</p>
        </div>
      {/if}
    {/if}

  <!-- ── General tab ───────────────────────────────────────────────────── -->
  {:else if tab === 'general'}
    {#if wsLoading}
      <div class="flex items-center gap-2 text-muted-foreground py-6">
        <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
        Loading…
      </div>
    {:else}
      <div class="bg-card border border-border rounded-lg p-5 flex flex-col gap-4">
        <div>
          <label class="block text-[12px] font-semibold text-muted-foreground uppercase tracking-wide mb-1.5" for="ws-name">Name</label>
          <input
            id="ws-name"
            class="w-full text-[13px] border border-border rounded-md px-3 py-1.5 bg-background focus:outline-none focus:ring-1 focus:ring-primary"
            type="text"
            bind:value={wsForm.name}
            disabled={!canManage}
          />
        </div>

        <div>
          <label class="block text-[12px] font-semibold text-muted-foreground uppercase tracking-wide mb-1.5" for="ws-purpose">Purpose</label>
          <select
            id="ws-purpose"
            class="w-full text-[13px] border border-border rounded-md px-3 py-1.5 bg-background focus:outline-none focus:ring-1 focus:ring-primary"
            bind:value={wsForm.purpose}
            disabled={!canManage}
          >
            {#each PURPOSES as p}
              <option value={p}>{p}</option>
            {/each}
          </select>
        </div>

        <div>
          <label class="block text-[12px] font-semibold text-muted-foreground uppercase tracking-wide mb-1.5" for="ws-desc">Description</label>
          <textarea
            id="ws-desc"
            class="w-full text-[13px] border border-border rounded-md px-3 py-1.5 bg-background focus:outline-none focus:ring-1 focus:ring-primary resize-y min-h-[80px]"
            bind:value={wsForm.description}
            disabled={!canManage}
          ></textarea>
        </div>

        {#if wsError}
          <div class="text-[12px] text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-1.5">{wsError}</div>
        {/if}
        {#if wsSuccess}
          <div class="text-[12px] text-green-700 bg-green-50 border border-green-200 rounded-md px-3 py-1.5">Changes saved.</div>
        {/if}

        {#if canManage}
          <div class="flex justify-end">
            <button
              class="px-4 py-1.5 bg-primary text-primary-foreground text-[13px] font-medium rounded-md hover:bg-primary/90 transition-colors disabled:opacity-50"
              disabled={wsSaving}
              on:click={saveGeneral}
            >
              {wsSaving ? 'Saving…' : 'Save changes'}
            </button>
          </div>
        {/if}
      </div>

      <!-- Export -->
      <div class="bg-card border border-border rounded-lg p-5">
        <div class="text-[12px] font-semibold uppercase tracking-wide text-muted-foreground mb-1">Export model</div>
        <p class="text-[12px] text-muted-foreground mb-3">Download the full workspace model including all elements, relationships, and diagram layouts.</p>
        <div class="flex gap-2">
          <a
            href="/api/v1/workspaces/{wsId}/export/aoef"
            download
            class="inline-flex items-center gap-1.5 px-3 py-1.5 text-[13px] font-medium border border-border rounded-md hover:border-primary hover:text-primary transition-colors"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="size-3.5" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clip-rule="evenodd"/>
            </svg>
            AOEF (XML)
          </a>
          <a
            href="/api/v1/workspaces/{wsId}/export/ajx"
            download
            class="inline-flex items-center gap-1.5 px-3 py-1.5 text-[13px] font-medium border border-border rounded-md hover:border-primary hover:text-primary transition-colors"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="size-3.5" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clip-rule="evenodd"/>
            </svg>
            AJX (JSON)
          </a>
        </div>
      </div>

      <!-- Danger zone — owners only -->
      {#if isOwner}
        <div class="border border-destructive/40 rounded-lg overflow-hidden mt-2">
          <div class="bg-destructive/5 border-b border-destructive/40 px-5 py-3 flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="size-4 text-destructive flex-shrink-0" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
            </svg>
            <span class="text-[12px] font-semibold uppercase tracking-wide text-destructive">Danger zone</span>
          </div>
          <div class="p-5">
            <div class="flex items-start justify-between gap-4">
              <div>
                <div class="text-[13px] font-medium">Delete this workspace</div>
                <p class="text-[12px] text-muted-foreground mt-0.5">
                  Permanently removes all elements, relationships, diagrams, members, and history.
                  This action <strong>cannot be undone</strong>.
                </p>
              </div>
            </div>
            <div class="mt-4 pt-4 border-t border-destructive/20">
              <label class="block text-[12px] text-muted-foreground mb-1.5" for="delete-confirm">
                Type <strong class="text-foreground">{ws.name}</strong> to confirm
              </label>
              <div class="flex gap-2">
                <input
                  id="delete-confirm"
                  class="flex-1 text-[13px] border rounded-md px-3 py-1.5 bg-background focus:outline-none focus:ring-1
                    {deleteReady ? 'border-destructive focus:ring-destructive' : 'border-border focus:ring-primary'}"
                  type="text"
                  placeholder={ws.name}
                  bind:value={deleteConfirm}
                  disabled={deleting}
                />
                <button
                  class="px-4 py-1.5 text-[13px] font-medium rounded-md transition-colors
                    {deleteReady
                      ? 'bg-destructive text-destructive-foreground hover:bg-destructive/90'
                      : 'bg-muted text-muted-foreground cursor-not-allowed opacity-50'}"
                  disabled={!deleteReady || deleting}
                  on:click={deleteWorkspace}
                >
                  {deleting ? 'Deleting…' : 'Delete workspace'}
                </button>
              </div>
              {#if deleteError}
                <div class="mt-2 text-[12px] text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-1.5">{deleteError}</div>
              {/if}
            </div>
          </div>
        </div>
      {/if}
    {/if}
  {/if}
</div>
