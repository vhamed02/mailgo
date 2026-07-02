'use client'

import { useEffect, useState } from 'react'
import { Menu, X } from 'lucide-react'
import { Container } from '../container'
import { Logo } from '../logo'
import { ButtonLink } from '../button-link'
import { cn } from '@/lib/utils'

const NAV_LINKS = [
  { label: 'Features', href: '#features' },
  { label: 'Product', href: '#product' },
  { label: 'Customers', href: '#customers' },
  { label: 'Security', href: '#security' },
]

export function Navbar() {
  const [open, setOpen] = useState(false)
  const [scrolled, setScrolled] = useState(false)

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  useEffect(() => {
    document.body.style.overflow = open ? 'hidden' : ''
    return () => {
      document.body.style.overflow = ''
    }
  }, [open])

  return (
    <header
      className={cn(
        'sticky top-0 z-50 w-full transition-colors',
        scrolled
          ? 'border-b border-slate-200/80 bg-white/85 backdrop-blur-md'
          : 'border-b border-transparent bg-transparent'
      )}
    >
      <Container className="flex h-16 items-center justify-between lg:h-18">
        <Logo />

        <nav className="hidden items-center gap-8 md:flex" aria-label="Primary">
          {NAV_LINKS.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="text-sm font-medium text-slate-600 transition-colors hover:text-slate-900"
            >
              {link.label}
            </a>
          ))}
        </nav>

        <div className="hidden items-center gap-2 md:flex">
          <ButtonLink href="/auth/login" variant="ghost" size="md">
            Sign in
          </ButtonLink>
          <ButtonLink href="/auth/register" variant="primary" size="md">
            Get started
          </ButtonLink>
        </div>

        <button
          type="button"
          className="inline-flex h-10 w-10 items-center justify-center rounded-lg text-slate-700 hover:bg-slate-100 md:hidden"
          aria-label={open ? 'Close menu' : 'Open menu'}
          aria-expanded={open}
          onClick={() => setOpen((v) => !v)}
        >
          {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </Container>

      {open ? (
        <div className="md:hidden">
          <div className="space-y-1 border-t border-slate-200 bg-white px-4 py-4">
            {NAV_LINKS.map((link) => (
              <a
                key={link.href}
                href={link.href}
                onClick={() => setOpen(false)}
                className="block rounded-lg px-3 py-2.5 text-base font-medium text-slate-700 hover:bg-slate-50"
              >
                {link.label}
              </a>
            ))}
            <div className="grid grid-cols-2 gap-2 pt-3">
              <ButtonLink
                href="/auth/login"
                variant="secondary"
                size="md"
                onClick={() => setOpen(false)}
              >
                Sign in
              </ButtonLink>
              <ButtonLink
                href="/auth/register"
                variant="primary"
                size="md"
                onClick={() => setOpen(false)}
              >
                Get started
              </ButtonLink>
            </div>
          </div>
        </div>
      ) : null}
    </header>
  )
}
