import { Star, Quote } from 'lucide-react'
import { Container } from '../container'
import { Section } from '../section'
import { SectionHeading } from '../section-heading'
import { cn } from '@/lib/utils'

type Testimonial = {
  quote: string
  name: string
  role: string
  initials: string
  color: string
  featured?: boolean
}

const TESTIMONIALS: Testimonial[] = [
  {
    quote:
      'We migrated 400 mailboxes to MailGo over a weekend and never looked back. Deliverability jumped and our support tickets about email all but disappeared.',
    name: 'Sarah Patel',
    role: 'Head of IT, Northwind',
    initials: 'SP',
    color: 'bg-indigo-500',
    featured: true,
  },
  {
    quote: 'The webmail is genuinely fast. Our team stopped asking for a desktop client within a week.',
    name: 'Diego Romero',
    role: 'Ops Lead, Globex',
    initials: 'DR',
    color: 'bg-rose-500',
  },
  {
    quote: 'SPF, DKIM and DMARC were configured for us automatically. That alone saved me a full day of DNS work.',
    name: 'Aiko Tanaka',
    role: 'Founder, Initech',
    initials: 'AT',
    color: 'bg-emerald-500',
  },
  {
    quote: 'Provisioning a new joiner\u2019s mailbox takes seconds. The admin dashboard is exactly what we wanted.',
    name: 'Marcus Webb',
    role: 'CTO, Hooli',
    initials: 'MW',
    color: 'bg-sky-500',
  },
  {
    quote: 'Inbox placement analytics finally give us visibility we never had with our previous provider.',
    name: 'Lena Fischer',
    role: 'Growth, Vandelay',
    initials: 'LF',
    color: 'bg-amber-500',
  },
]

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
  return (
    <Section id="customers" className="bg-white">
      <Container>
        <SectionHeading
          eyebrow="Customers"
          title="Loved by teams who rely on email"
          description="From two-person startups to organizations with hundreds of mailboxes, teams trust MailGo to keep their email fast, secure and delivered."
        />

        <div className="mt-14 grid grid-cols-1 gap-6 lg:mt-16 lg:grid-cols-3">
          {TESTIMONIALS.map((t) => (
            <TestimonialCard key={t.name} t={t} />
          ))}
        </div>
      </Container>
    </Section>
  )
}
