import { writable } from 'svelte/store';

// Increment to signal that saved views were modified and sidebars should refresh.
export const savedViewsRevision = writable(0);
