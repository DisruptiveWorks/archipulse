import { writable } from 'svelte/store';

// Incremented each time a model is successfully imported.
// Components that display model data subscribe to this to trigger a reload.
export const importRevision = writable(0);
