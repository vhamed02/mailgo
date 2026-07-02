import { Globe, Users, ShieldCheck, Zap, BarChart3, Lock } from 'lucide-react'
import { Container } from '../container'
import { Section } from '../section'
import { SectionHeading } from '../section-heading'
import { cn } from '@/lib/utils'

type Feature = {
  icon: React.ElementType
  title: string
  description: string
  accent: string
}

const FEATURES: Feature[] = [
  {
    icon: Globe,
    title: 'Custom domains',
    description:
      'Bring your own domain and spin up branded mailboxes like hello@yourcompany.com in minutes, with automatic DNS verification.',
    accent: 'indigo',
  },
  {
    icon: Users,
    title: 'Team management',
    description:
      'Provision mailboxes, assign roles, and delegate admins. Invite your whole team and control access from one dashboard.',
    accent: 'blue',
  },
  {
    icon: ShieldCheck,
    title: 'Enterprise security',
    description:
      'SPF, DKIM and DMARC come preconfigured. Enforce 2FA, audit access, and keep phishing out of the inbox.',
    accent: 'violet',
  },
  {
    icon: Zap,
    title: 'Lightning webmail',
    description:
      'A fast, keyboard-friendly webmail with threaded conversations, instant search and a clean, distraction-free composer.',
    accent: 'sky',
  },
  {
    icon: BarChart3,
    title: 'Deliverability toolkit',
    description:
      'Real-time reputation monitoring, bounce analytics and queue insights so your messages actually land in the inbox.',
    accent: 'cyan',
  },
  {
    icon: Lock,
    title: 'Encryption at rest',
    description:
      'Mailboxes are encrypted at rest with per-tenant keys, plus optional end-to-end encryption for sensitive conversations.',
    accent: 'teal',
  },
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
  return (
    <Section id="features" className="bg-white">
      <Container>
        <SectionHeading
          eyebrow="Features"
          title="Everything you need to run business email"
          description="MailGo brings hosting, webmail, admin controls and deliverability tools together, so your team can stop stitching tools together and just send."
        />

        <div className="mt-14 grid grid-cols-1 gap-6 sm:grid-cols-2 lg:mt-16 lg:grid-cols-3">
          {FEATURES.map((feature) => {
            const Icon = feature.icon
            const accent = ACCENT_MAP[feature.accent]
            return (
              <div
                key={feature.title}
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
                    {feature.title}
                  </h3>
                  <p className="mt-2 text-sm leading-relaxed text-slate-600">
                    {feature.description}
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
