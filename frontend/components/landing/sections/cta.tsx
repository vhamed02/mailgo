import { ArrowRight } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { Container } from '../container'
import { ButtonLink } from '../button-link'

export function CTA() {
  const t = useTranslations('cta')
  return (
    <section id="cta" className="scroll-mt-24 bg-white py-20 sm:py-24 lg:py-28">
      <Container>
        <div className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-indigo-600 via-blue-600 to-indigo-700 px-6 py-16 text-center shadow-2xl shadow-indigo-600/25 sm:px-12 sm:py-20">
          <div className="pointer-events-none absolute inset-0 bg-grid-slate opacity-20 [mask-image:radial-gradient(ellipse_at_center,black,transparent_75%)]" />
          <div className="pointer-events-none absolute -right-16 -top-16 h-64 w-64 rounded-full bg-white/10 blur-2xl" />
          <div className="pointer-events-none absolute -bottom-20 -left-10 h-64 w-64 rounded-full bg-indigo-300/30 blur-3xl" />

          <div className="relative mx-auto max-w-2xl">
            <h2 className="text-3xl font-bold tracking-tight text-white sm:text-4xl lg:text-5xl">
              {t('title')}
            </h2>
            <p className="mx-auto mt-4 max-w-xl text-lg leading-relaxed text-indigo-100">
              {t('description')}
            </p>
            <div className="mt-9 flex flex-col items-center justify-center gap-3 sm:flex-row">
              <ButtonLink
                href="/auth/register"
                variant="light"
                size="xl"
                className="w-full sm:w-auto"
              >
                {t('ctaPrimary')}
                <ArrowRight className="h-5 w-5" />
              </ButtonLink>
              <ButtonLink
                href="/auth/login"
                variant="ghost"
                size="xl"
                className="w-full text-white hover:bg-white/10 sm:w-auto"
              >
                {t('signIn')}
              </ButtonLink>
            </div>
            <p className="mt-6 text-sm text-indigo-200">
              {t('trust')}
            </p>
          </div>
        </div>
      </Container>
    </section>
  )
}
