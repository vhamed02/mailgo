import { useTranslations } from 'next-intl'
import { Container } from '../container'

const LOGOS = ['Acme', 'Globex', 'Initech', 'Umbrella', 'Hooli', 'Soylent', 'Vandelay', 'Pied Piper']

export function LogoCloud() {
  const t = useTranslations('logos')
  return (
    <section className="border-y border-slate-200/70 bg-slate-50/60 py-12 sm:py-14">
      <Container>
        <p className="text-center text-xs font-semibold uppercase tracking-widest text-slate-400">
          {t('trusted')}
        </p>
        <div className="mt-8 grid grid-cols-2 items-center gap-x-6 gap-y-8 sm:grid-cols-4 lg:grid-cols-8">
          {LOGOS.map((logo) => (
            <div
              key={logo}
              className="flex items-center justify-center text-lg font-bold tracking-tight text-slate-400 grayscale transition-all hover:text-slate-600"
              aria-label={logo}
            >
              {logo}
            </div>
          ))}
        </div>
      </Container>
    </section>
  )
}
