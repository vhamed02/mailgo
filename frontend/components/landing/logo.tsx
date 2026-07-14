import Link from 'next/link'
import { cn } from '@/lib/utils'

type LogoProps = {
  className?: string
  href?: string
  showWordmark?: boolean
}

export function Logo({ className, href = '/', showWordmark = true }: LogoProps) {
  return (
    <Link
      href={href}
      className={cn('group inline-flex items-center gap-2.5', className)}
      aria-label="Mailbox home"
    >
      <span className="relative inline-flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-indigo-600 to-blue-600 shadow-lg shadow-indigo-600/25">
        <svg
          viewBox="0 0 24 24"
          fill="none"
          className="h-5 w-5 text-white"
          aria-hidden="true"
        >
          <path
            d="M4 7.5A1.5 1.5 0 0 1 5.5 6h13A1.5 1.5 0 0 1 20 7.5v9a1.5 1.5 0 0 1-1.5 1.5h-13A1.5 1.5 0 0 1 4 16.5v-9Z"
            stroke="currentColor"
            strokeWidth="1.6"
            strokeLinejoin="round"
          />
          <path
            d="m4.6 7.4 6.2 4.2a2 2 0 0 0 2.4 0l6.2-4.2"
            stroke="currentColor"
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </span>
      {showWordmark ? (
        <span className="text-lg font-bold tracking-tight text-slate-900">
          Mail<span className="text-indigo-600">box</span>
        </span>
      ) : null}
    </Link>
  )
}
