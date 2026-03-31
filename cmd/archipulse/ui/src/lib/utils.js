import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

/** Merge Tailwind classes, resolving conflicts with tailwind-merge. */
export function cn(...inputs) {
  return twMerge(clsx(inputs));
}
