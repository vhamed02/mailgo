import { Check, ArrowRight, ShieldCheck, KeyRound, FileLock, Activity } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { Container } from '../container'
import { Section } from '../section'
import { cn } from '@/lib/utils'

type FeatureRowProps = {
  id?: string
  eyebrow: string
  title: string
  description: string
  bullets: string[]
  reverse?: boolean
  visual: React.ReactNode
}

function FeatureRow({
  id,
  eyebrow,
  title,
  description,
  bullets,
  reverse,
  visual,
}: FeatureRowProps) {
  const t = useTranslations('product')
  return (
    <div id={id} className="grid grid-cols-1 items-center gap-10 lg:grid-cols-2 lg:gap-16 scroll-mt-24">
      <div className={cn(reverse ? 'lg:order-2' : 'lg:order-1')}>
        <p className="text-sm font-semibold uppercase tracking-wider text-indigo-600">
          {eyebrow}
        </p>
        <h3 className="mt-3 text-2xl font-bold tracking-tight text-slate-900 sm:text-3xl">
          {title}
        </h3>
        <p className="mt-4 text-base leading-relaxed text-slate-600 sm:text-lg">
          {description}
        </p>
        <ul className="mt-6 space-y-3">
          {bullets.map((bullet) => (
            <li key={bullet} className="flex items-start gap-3 text-sm text-slate-700 sm:text-base">
              <span className="mt-0.5 inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-indigo-100">
                <Check className="h-3.5 w-3.5 text-indigo-600" />
              </span>
              {bullet}
            </li>
          ))}
        </ul>
        <a
          href="#cta"
          className="mt-7 inline-flex items-center gap-1.5 text-sm font-semibold text-indigo-600 hover:text-indigo-700"
        >
          {t('learnMore')}
          <ArrowRight className="h-4 w-4" />
        </a>
      </div>
      <div className={cn('lg:order-1', reverse ? 'lg:order-1' : 'lg:order-2')}>
        {visual}
      </div>
    </div>
  )
}

function PanelFrame({ children }: { children: React.ReactNode }) {
  return (
    <div className="overflow-hidden rounded-2xl border border-slate-200/80 bg-white shadow-card ring-1 ring-slate-900/5">
      {children}
    </div>
  )
}

function WebmailVisual() {
  const thread = [
    { from: 'Maya Chen', initials: 'MC', color: 'bg-rose-500', time: '9:41', body: 'Pushed the latest webmail changes — thread view is live. Mind reviewing before we ship to staging?' },
    { from: 'You', initials: 'YO', color: 'bg-indigo-500', time: '9:44', body: 'Looks great. Let\u2019s ship it. ', me: true },
    { from: 'Maya Chen', initials: 'MC', color: 'bg-rose-500', time: '9:46', body: 'Done. I\u2019ll tag the release.' },
  ]
  return (
    <PanelFrame>
      <div className="flex items-center justify-between border-b border-slate-100 px-5 py-3">
        <div>
          <p className="text-sm font-semibold text-slate-900">Re: Q3 roadmap sync</p>
          <p className="text-xs text-slate-400">3 messages · thread</p>
        </div>
        <div className="flex -space-x-2">
          <span className="flex h-7 w-7 items-center justify-center rounded-full bg-rose-500 text-[10px] font-semibold text-white ring-2 ring-white">MC</span>
          <span className="flex h-7 w-7 items-center justify-center rounded-full bg-indigo-500 text-[10px] font-semibold text-white ring-2 ring-white">YO</span>
        </div>
      </div>
      <div className="space-y-4 p-5">
        {thread.map((m) => (
          <div key={m.from + m.time} className={cn('flex gap-3', m.me && 'flex-row-reverse text-right')}>
            <span className={cn('flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-[11px] font-semibold text-white', m.color)}>
              {m.initials}
            </span>
            <div className={cn('max-w-[80%] rounded-2xl px-4 py-2.5 text-sm', m.me ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-700')}>
              <p>{m.body}</p>
              <p className={cn('mt-1 text-[11px]', m.me ? 'text-indigo-200' : 'text-slate-400')}>{m.time}</p>
            </div>
          </div>
        ))}
        <div className="flex items-center gap-2 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-400">
          Write a reply…
        </div>
      </div>
    </PanelFrame>
  )
}

function DashboardVisual() {
  const stats = [
    { label: 'Mailboxes', value: '128', delta: '+12' },
    { label: 'Storage used', value: '412 GB', delta: '+8%' },
    { label: 'Deliverability', value: '99.2%', delta: '+0.4' },
  ]
  const domains = [
    { domain: 'acme.com', mailboxes: 64, status: 'Healthy' },
    { domain: 'mail.acme.com', mailboxes: 38, status: 'Healthy' },
    { domain: 'partners.acme.com', mailboxes: 26, status: 'DNS review' },
  ]
  return (
    <PanelFrame>
      <div className="border-b border-slate-100 px-5 py-3">
        <p className="text-sm font-semibold text-slate-900">Organization overview</p>
        <p className="text-xs text-slate-400">acme.com · updated just now</p>
      </div>
      <div className="grid grid-cols-3 gap-3 p-5">
        {stats.map((s) => (
          <div key={s.label} className="rounded-xl border border-slate-200 bg-white p-3">
            <p className="text-xs text-slate-500">{s.label}</p>
            <p className="mt-1 text-lg font-bold text-slate-900">{s.value}</p>
            <p className="text-[11px] font-medium text-emerald-600">{s.delta}</p>
          </div>
        ))}
      </div>
      <div className="px-5 pb-5">
        <div className="overflow-hidden rounded-xl border border-slate-200">
          <div className="grid grid-cols-[1fr_auto_auto] gap-2 border-b border-slate-100 bg-slate-50 px-3 py-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">
            <span>Domain</span>
            <span className="text-right">Mailboxes</span>
            <span className="text-right">Status</span>
          </div>
          {domains.map((d) => (
            <div key={d.domain} className="grid grid-cols-[1fr_auto_auto] gap-2 border-b border-slate-50 px-3 py-2.5 text-sm last:border-0">
              <span className="font-medium text-slate-700">{d.domain}</span>
              <span className="text-right text-slate-500">{d.mailboxes}</span>
              <span className="text-right">
                <span className={cn('inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium', d.status === 'Healthy' ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700')}>
                  {d.status}
                </span>
              </span>
            </div>
          ))}
        </div>
      </div>
    </PanelFrame>
  )
}

function SecurityVisual() {
  const checks = [
    { label: 'SPF', icon: ShieldCheck, value: 'Aligned', ok: true },
    { label: 'DKIM', icon: KeyRound, value: 'Signing 1024-bit', ok: true },
    { label: 'DMARC', icon: FileLock, value: 'p=quarantine', ok: true },
    { label: 'Reputation', icon: Activity, value: 'Excellent', ok: true },
  ]
  return (
    <PanelFrame>
      <div className="border-b border-slate-100 px-5 py-3">
        <p className="text-sm font-semibold text-slate-900">Deliverability & security</p>
        <p className="text-xs text-slate-400">acme.com</p>
      </div>
      <div className="grid grid-cols-2 gap-3 p-5">
        {checks.map((c) => {
          const Icon = c.icon
          return (
            <div key={c.label} className="flex items-center gap-3 rounded-xl border border-slate-200 bg-white p-3">
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600">
                <Icon className="h-5 w-5" />
              </span>
              <div>
                <p className="text-xs text-slate-500">{c.label}</p>
                <p className="text-sm font-semibold text-slate-900">{c.value}</p>
              </div>
              <span className="ml-auto inline-flex h-5 w-5 items-center justify-center rounded-full bg-emerald-500 text-white">
                <Check className="h-3 w-3" />
              </span>
            </div>
          )
        })}
      </div>
      <div className="px-5 pb-5">
        <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
          <div className="mb-1 flex items-center justify-between text-xs text-slate-500">
            <span>Inbox placement (30d)</span>
            <span className="font-semibold text-slate-700">99.2%</span>
          </div>
          <div className="flex h-16 items-end gap-1.5">
            {[40, 55, 48, 62, 58, 70, 66, 78, 74, 82, 80, 88, 92, 90, 96].map((h, i) => (
              <div
                key={i}
                className="flex-1 rounded-t bg-gradient-to-t from-indigo-500 to-blue-400"
                style={{ height: `${h}%` }}
              />
            ))}
          </div>
        </div>
      </div>
    </PanelFrame>
  )
}

export function ValueProposition() {
  const t = useTranslations('product')
  return (
    <Section id="product" className="bg-slate-50/60">
      <Container>
        <div className="mx-auto max-w-3xl text-center">
          <p className="text-sm font-semibold uppercase tracking-wider text-indigo-600">
            {t('eyebrow')}
          </p>
          <h2 className="mt-3 text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl lg:text-5xl">
            {t('title')}
          </h2>
          <p className="mt-4 text-lg leading-relaxed text-slate-600">
            {t('description')}
          </p>
        </div>

        <div className="mt-16 space-y-20 lg:space-y-28">
          <FeatureRow
            eyebrow={t('rows.webmail.eyebrow')}
            title={t('rows.webmail.title')}
            description={t('rows.webmail.description')}
            bullets={[
              t('rows.webmail.bullet1'),
              t('rows.webmail.bullet2'),
              t('rows.webmail.bullet3'),
              t('rows.webmail.bullet4'),
            ]}
            visual={<WebmailVisual />}
          />
          <FeatureRow
            eyebrow={t('rows.admin.eyebrow')}
            title={t('rows.admin.title')}
            description={t('rows.admin.description')}
            bullets={[
              t('rows.admin.bullet1'),
              t('rows.admin.bullet2'),
              t('rows.admin.bullet3'),
              t('rows.admin.bullet4'),
            ]}
            reverse
            visual={<DashboardVisual />}
          />
          <FeatureRow
            id="security"
            eyebrow={t('rows.security.eyebrow')}
            title={t('rows.security.title')}
            description={t('rows.security.description')}
            bullets={[
              t('rows.security.bullet1'),
              t('rows.security.bullet2'),
              t('rows.security.bullet3'),
              t('rows.security.bullet4'),
            ]}
            visual={<SecurityVisual />}
          />
        </div>
      </Container>
    </Section>
  )
}
