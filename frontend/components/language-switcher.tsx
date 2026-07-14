'use client'

import { useEffect, useRef, useState, useTransition } from 'react'
import { useLocale } from 'next-intl'
import { useRouter } from 'next/navigation'
import { Check, Globe, ChevronDown } from 'lucide-react'
import { locales, localeNames, localeShort, LOCALE_COOKIE, type Locale } from '@/i18n/config'
import { cn } from '@/lib/utils'

type Props = {
  /** Visual style: 'light' for light backgrounds, 'contrast' for dark/coloured backgrounds. */
  variant?: 'light' | 'contrast'
  className?: string
}

export function LanguageSwitcher({ variant = 'light', className }: Props) {
  const active = useLocale() as Locale
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [pending, startTransition] = useTransition()
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onClick)
    return () => document.removeEventListener('mousedown', onClick)
  }, [])

  const select = (next: Locale) => {
    setOpen(false)
    if (next === active) return
    // Persist for a year; the server reads this cookie to pick messages.
    document.cookie = `${LOCALE_COOKIE}=${next};path=/;max-age=31536000;samesite=lax`
    document.documentElement.lang = next
    startTransition(() => router.refresh())
  }

  return (
    <div ref={ref} className={cn('relative', className)}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-label="Change language"
        disabled={pending}
        className={cn(
          'inline-flex items-center gap-1.5 rounded-lg px-2.5 py-2 text-sm font-medium transition-colors disabled:opacity-60',
          variant === 'contrast'
            ? 'text-white/90 hover:bg-white/10'
            : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900'
        )}
      >
        <Globe className="h-4 w-4" />
        <span>{localeShort[active]}</span>
        <ChevronDown className={cn('h-3.5 w-3.5 transition-transform', open && 'rotate-180')} />
      </button>

      {open ? (
        <ul
          role="listbox"
          className="absolute right-0 z-50 mt-2 w-44 overflow-hidden rounded-xl border border-slate-200 bg-white py-1 shadow-lg shadow-slate-900/10"
        >
          {locales.map((code) => (
            <li key={code}>
              <button
                type="button"
                role="option"
                aria-selected={code === active}
                onClick={() => select(code)}
                className={cn(
                  'flex w-full items-center justify-between px-3 py-2 text-left text-sm transition-colors hover:bg-slate-50',
                  code === active ? 'font-semibold text-indigo-600' : 'text-slate-700'
                )}
              >
                {localeNames[code]}
                {code === active ? <Check className="h-4 w-4" /> : null}
              </button>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  )
}
