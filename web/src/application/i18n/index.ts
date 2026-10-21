import { derived, get, writable } from 'svelte/store';
import type { StoragePort } from '../../ports/browser_port';
import { en } from './en';
import { fr } from './fr';

export type Lang = 'en' | 'fr';
export type Translate = (key: string, params?: Record<string, string | number>) => string;

const dictionaries: Record<Lang, Record<string, string>> = { en, fr };
const STORAGE_KEY = 'fana.lang';

let storage: StoragePort | null = null;

function detectLang(): Lang {
  const saved = storage?.read(STORAGE_KEY);
  if (saved === 'en' || saved === 'fr') return saved;
  if (typeof navigator !== 'undefined' && navigator.language.toLowerCase().startsWith('fr')) return 'fr';
  return 'en';
}

export const lang = writable<Lang>('en');

/** Wires persistence and the html lang attribute to the injected storage port. */
export function initI18n(port: StoragePort): void {
  storage = port;
  lang.set(detectLang());
  lang.subscribe((value) => {
    if (typeof document !== 'undefined') document.documentElement.lang = value;
    storage?.write(STORAGE_KEY, value);
  });
}

export function setLang(value: Lang): void {
  lang.set(value);
}

export function toggleLang(): void {
  lang.set(get(lang) === 'en' ? 'fr' : 'en');
}

export const tr = derived(lang, ($lang): Translate => {
  return (key, params) => {
    let text = dictionaries[$lang][key] ?? dictionaries.en[key] ?? key;
    if (params) {
      for (const [name, value] of Object.entries(params)) {
        text = text.replaceAll(`{${name}}`, String(value));
      }
    }
    return text;
  };
});

export function errorText(translate: Translate, code: string): string {
  const key = `error.${code}`;
  const value = translate(key);
  return value === key ? translate('error.generic') : value;
}

export function noticeText(
  translate: Translate,
  code: string,
  params: Record<string, string>
): string {
  const value = translate(code, params);
  return value === code ? '' : value;
}
