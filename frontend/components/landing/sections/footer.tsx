import { Github, Twitter, Linkedin } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { Container } from '../container'
import { Logo } from '../logo'

type FooterColumn = {
  titleKey: string
  links: { labelKey: string; href: string }[]
}

const COLUMNS: FooterColumn[] = [
  {
    titleKey: 'product',
    links: [
      { labelKey: 'features', href: '#features' },
      { labelKey: 'webmail', href: '#product' },
      { labelKey: 'security', href: '#security' },
      { labelKey: 'pricing', href: '#cta' },
      { labelKey: 'changelog', href: '#' },
    ],
  },
  {
    titleKey: 'company',
    links: [
      { labelKey: 'about', href: '#' },
      { labelKey: 'customers', href: '#customers' },
      { labelKey: 'careers', href: '#' },
      { labelKey: 'blog', href: '#' },
      { labelKey: 'contact', href: '#' },
    ],
  },
  {
    titleKey: 'resources',
    links: [
      { labelKey: 'documentation', href: '#' },
      { labelKey: 'apiReference', href: '#' },
      { labelKey: 'guides', href: '#' },
      { labelKey: 'status', href: '#' },
      { labelKey: 'support', href: '#' },
    ],
  },
  {
    titleKey: 'legal',
    links: [
      { labelKey: 'privacy', href: '#' },
      { labelKey: 'terms', href: '#' },
      { labelKey: 'security', href: '#' },
      { labelKey: 'dpa', href: '#' },
    ],
  },
]

const SOCIALS = [
  { label: 'GitHub', href: '#', icon: Github },
  { label: 'X', href: '#', icon: Twitter },
  { label: 'LinkedIn', href: '#', icon: Linkedin },
]

export function Footer() {
  const t = useTranslations('footer')
  const year = new Date().getFullYear()
  return (
    <footer className="border-t border-slate-200 bg-slate-50">
      <Container className="py-14 lg:py-16">
        <div className="grid grid-cols-2 gap-10 md:grid-cols-3 lg:grid-cols-[1.6fr_repeat(4,1fr)]">
          <div className="col-span-2 md:col-span-3 lg:col-span-1">
            <Logo />
            <p className="mt-4 max-w-xs text-sm leading-relaxed text-slate-500">
              {t('description')}
            </p>
            <div className="mt-5 flex items-center gap-2">
              {SOCIALS.map((s) => {
                const Icon = s.icon
                return (
                  <a
                    key={s.label}
                    href={s.href}
                    aria-label={s.label}
                    className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 bg-white text-slate-500 transition-colors hover:border-slate-300 hover:text-slate-900"
                  >
                    <Icon className="h-4 w-4" />
                  </a>
                )
              })}
            </div>
          </div>

          {COLUMNS.map((col) => (
            <div key={col.titleKey}>
              <h3 className="text-sm font-semibold text-slate-900">{t(`cols.${col.titleKey}`)}</h3>
              <ul className="mt-4 space-y-3">
                {col.links.map((link) => (
                  <li key={link.labelKey}>
                    <a
                      href={link.href}
                      className="text-sm text-slate-500 transition-colors hover:text-slate-900"
                    >
                      {t(`links.${link.labelKey}`)}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="mt-12 flex flex-col items-center justify-between gap-4 border-t border-slate-200 pt-6 sm:flex-row">
          <p className="text-sm text-slate-500">
            {t('rights', { year })}
          </p>
          <div className="flex items-center gap-2 text-sm text-slate-500">
            <span className="inline-flex h-2 w-2 rounded-full bg-emerald-500" />
            {t('operational')}
          </div>
        </div>
      </Container>
    </footer>
  )
}
