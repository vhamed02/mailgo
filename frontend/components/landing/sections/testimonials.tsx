import { Star, Quote } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { Container } from '../container'
import { Section } from '../section'
import { SectionHeading } from '../section-heading'
import { cn } from '@/lib/utils'

type Testimonial = {
  key: string
  quote: string
  name: string
  role: string
  initials: string
  color: string
  featured?: boolean
}

const TESTIMONIAL_META = [
  { key: 'one', name: 'Sarah Patel', initials: 'SP', color: 'bg-indigo-500', featured: true },
  { key: 'two', name: 'Diego Romero', initials: 'DR', color: 'bg-rose-500' },
  { key: 'three', name: 'Aiko Tanaka', initials: 'AT', color: 'bg-emerald-500' },
  { key: 'four', name: 'Marcus Webb', initials: 'MW', color: 'bg-sky-500' },
  { key: 'five', name: 'Lena Fischer', initials: 'LF', color: 'bg-amber-500' },
] as const

function Stars() {
  return (
    <div className="flex items-center gap-0.5 text-amber-400" aria-label="5 out of 5 stars">
      {Array.from({ length: 5 }).map((_, i) => (
        <Star key={i} className="h-4 w-4 fill-current" />
      ))}
    </div>
  )
}

function TestimonialCard({ t }: { t: Testimonial }) {
  return (
    <figure
      className={cn(
        'flex h-full flex-col justify-between gap-6 rounded-2xl border border-slate-200 bg-white p-7 shadow-card transition-all duration-200 hover:-translate-y-1 hover:shadow-card-hover',
        t.featured && 'lg:col-span-2 lg:flex-row lg:items-center lg:gap-10'
      )}
    >
      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <Stars />
          <Quote className="h-7 w-7 text-slate-200" />
        </div>
        <blockquote
          className={cn(
            'text-slate-700',
            t.featured ? 'text-xl font-medium leading-relaxed lg:text-2xl' : 'text-base leading-relaxed'
          )}
        >
          “{t.quote}”
        </blockquote>
      </div>
      <figcaption className="flex items-center gap-3">
        <span className={cn('flex h-10 w-10 items-center justify-center rounded-full text-sm font-semibold text-white', t.color)}>
          {t.initials}
        </span>
        <div>
          <p className="text-sm font-semibold text-slate-900">{t.name}</p>
          <p className="text-sm text-slate-500">{t.role}</p>
        </div>
      </figcaption>
    </figure>
  )
}

export function Testimonials() {
  const t = useTranslations('testimonials')
  const testimonials: Testimonial[] = TESTIMONIAL_META.map((m) => ({
    ...m,
    quote: t(`items.${m.key}.quote`),
    role: t(`items.${m.key}.role`),
    featured: 'featured' in m ? m.featured : undefined,
  }))
  return (
    <Section id="customers" className="bg-white">
      <Container>
        <SectionHeading
          eyebrow={t('eyebrow')}
          title={t('title')}
          description={t('description')}
        />

        <div className="mt-14 grid grid-cols-1 gap-6 lg:mt-16 lg:grid-cols-3">
          {testimonials.map((item) => (
            <TestimonialCard key={item.key} t={item} />
          ))}
        </div>
      </Container>
    </Section>
  )
}
