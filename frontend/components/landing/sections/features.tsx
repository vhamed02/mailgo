import { Globe, Users, ShieldCheck, Zap, BarChart3, Lock } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { Container } from '../container'
import { Section } from '../section'
import { SectionHeading } from '../section-heading'
import { cn } from '@/lib/utils'

type Feature = {
  icon: React.ElementType
  key: string
  accent: string
}

const FEATURES: Feature[] = [
  { icon: Globe, key: 'domains', accent: 'indigo' },
  { icon: Users, key: 'team', accent: 'blue' },
  { icon: ShieldCheck, key: 'securityFeat', accent: 'violet' },
  { icon: Zap, key: 'webmail', accent: 'sky' },
  { icon: BarChart3, key: 'deliverability', accent: 'cyan' },
  { icon: Lock, key: 'encryption', accent: 'teal' },
]

const ACCENT_MAP: Record<string, { bg: string; text: string }> = {
  indigo: { bg: 'bg-indigo-50', text: 'text-indigo-600' },
  blue: { bg: 'bg-blue-50', text: 'text-blue-600' },
  violet: { bg: 'bg-violet-50', text: 'text-violet-600' },
  sky: { bg: 'bg-sky-50', text: 'text-sky-600' },
  cyan: { bg: 'bg-cyan-50', text: 'text-cyan-600' },
  teal: { bg: 'bg-teal-50', text: 'text-teal-600' },
}

export function Features() {
  const t = useTranslations('features')
  return (
    <Section id="features" className="bg-white">
      <Container>
        <SectionHeading
          eyebrow={t('eyebrow')}
          title={t('title')}
          description={t('description')}
        />

        <div className="mt-14 grid grid-cols-1 gap-6 sm:grid-cols-2 lg:mt-16 lg:grid-cols-3">
          {FEATURES.map((feature) => {
            const Icon = feature.icon
            const accent = ACCENT_MAP[feature.accent]
            return (
              <div
                key={feature.key}
                className="group relative flex flex-col gap-5 rounded-2xl border border-slate-200 bg-white p-7 shadow-card transition-all duration-200 hover:-translate-y-1 hover:border-indigo-200 hover:shadow-card-hover"
              >
                <span
                  className={cn(
                    'inline-flex h-12 w-12 items-center justify-center rounded-xl ring-1 ring-inset ring-slate-900/5 transition-transform duration-200 group-hover:scale-105',
                    accent.bg
                  )}
                >
                  <Icon className={cn('h-6 w-6', accent.text)} />
                </span>
                <div>
                  <h3 className="text-lg font-semibold text-slate-900">
                    {t(`items.${feature.key}.title`)}
                  </h3>
                  <p className="mt-2 text-sm leading-relaxed text-slate-600">
                    {t(`items.${feature.key}.description`)}
                  </p>
                </div>
              </div>
            )
          })}
        </div>
      </Container>
    </Section>
  )
}
