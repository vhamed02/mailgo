// Language ordering for Mailbox: Armenian (default) → Russian → English.
export const locales = ['hy', 'ru', 'en'] as const
export type Locale = (typeof locales)[number]

export const defaultLocale: Locale = 'hy'

// Native names, shown in the language switcher in the same order as `locales`.
export const localeNames: Record<Locale, string> = {
  hy: 'Հայերեն',
  ru: 'Русский',
  en: 'English',
}

// Short codes for the compact switcher trigger.
export const localeShort: Record<Locale, string> = {
  hy: 'ՀԱՅ',
  ru: 'РУС',
  en: 'ENG',
}

export const LOCALE_COOKIE = 'NEXT_LOCALE'

export function isLocale(value: string | undefined | null): value is Locale {
  return !!value && (locales as readonly string[]).includes(value)
}
