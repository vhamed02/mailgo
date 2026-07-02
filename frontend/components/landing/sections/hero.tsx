import { ArrowRight, Check, Sparkles } from 'lucide-react'
import { Container } from '../container'
import { ButtonLink } from '../button-link'
import { Badge } from '../badge'

const TRUST_ITEMS = ['No credit card required', '14-day free trial', 'Cancel anytime']

export function Hero() {
  return (
    <section className="relative overflow-hidden">
      <div className="pointer-events-none absolute inset-0 -z-10">
        <div className="absolute inset-0 bg-grid-slate [mask-image:radial-gradient(ellipse_at_center,black,transparent_72%)]" />
        <div className="absolute left-1/2 top-[-12%] h-[34rem] w-[64rem] -translate-x-1/2 rounded-full bg-gradient-to-tr from-indigo-200/50 via-blue-200/40 to-transparent blur-3xl" />
        <div className="absolute right-[-10%] top-[20%] h-72 w-72 rounded-full bg-indigo-100/60 blur-3xl" />
      </div>

      <Container className="pb-16 pt-20 sm:pb-24 lg:pb-28 lg:pt-28">
        <div className="mx-auto max-w-3xl text-center">
          <a
            href="#product"
            className="animate-fade-up inline-flex"
            style={{ animationDelay: '60ms' }}
          >
            <Badge>
              <Sparkles className="h-3.5 w-3.5" />
              New · Webmail 2.0 is here
            </Badge>
          </a>

          <h1
            className="animate-fade-up mt-6 text-4xl font-bold tracking-tight text-slate-900 sm:text-6xl lg:text-7xl"
            style={{ animationDelay: '120ms' }}
          >
            Professional email hosting,
            <span className="text-gradient-brand"> built for modern teams</span>
          </h1>

          <p
            className="animate-fade-up mx-auto mt-6 max-w-2xl text-lg leading-relaxed text-slate-600 sm:text-xl"
            style={{ animationDelay: '180ms' }}
          >
            Spin up secure mailboxes on your own domain in minutes. MailGo pairs a
            lightning-fast webmail with the admin controls and deliverability tools
            your business actually needs.
          </p>

          <div
            className="animate-fade-up mt-9 flex flex-col items-center justify-center gap-3 sm:flex-row"
            style={{ animationDelay: '240ms' }}
          >
            <ButtonLink href="/auth/register" variant="primary" size="xl" className="w-full sm:w-auto">
              Get started free
              <ArrowRight className="h-5 w-5" />
            </ButtonLink>
            <ButtonLink href="#product" variant="secondary" size="xl" className="w-full sm:w-auto">
              See the product
            </ButtonLink>
          </div>

          <ul
            className="animate-fade-up mt-8 flex flex-wrap items-center justify-center gap-x-6 gap-y-2 text-sm text-slate-500"
            style={{ animationDelay: '300ms' }}
          >
            {TRUST_ITEMS.map((item) => (
              <li key={item} className="inline-flex items-center gap-1.5">
                <Check className="h-4 w-4 text-indigo-600" />
                {item}
              </li>
            ))}
          </ul>
        </div>

        <div
          className="animate-fade-up mx-auto mt-16 max-w-5xl"
          style={{ animationDelay: '360ms' }}
        >
          <HeroProductMock />
        </div>
      </Container>
    </section>
  )
}

function HeroProductMock() {
  const folders = [
    { name: 'Inbox', count: '24', active: true },
    { name: 'Sent', count: '182' },
    { name: 'Drafts', count: '3' },
    { name: 'Spam', count: '7' },
    { name: 'Trash', count: '' },
  ]

  const messages = [
    { from: 'Acme Procurement', subject: 'PO #4821 — confirmation needed', time: '9:42', unread: true, color: 'bg-indigo-500' },
    { from: 'Maya Chen', subject: 'Re: Q3 roadmap sync', time: '8:15', unread: true, color: 'bg-rose-500' },
    { from: 'GitHub', subject: '[mailgo/webmail] PR approved', time: 'Tue', unread: false, color: 'bg-slate-500' },
    { from: 'Stripe', subject: 'Your June payout arrived', time: 'Mon', unread: false, color: 'bg-emerald-500' },
  ]

  return (
    <div className="overflow-hidden rounded-2xl border border-slate-200/80 bg-white shadow-2xl shadow-slate-900/10 ring-1 ring-slate-900/5">
      <div className="flex items-center gap-2 border-b border-slate-200/80 bg-slate-50/80 px-4 py-3">
        <span className="h-3 w-3 rounded-full bg-rose-400" />
        <span className="h-3 w-3 rounded-full bg-amber-400" />
        <span className="h-3 w-3 rounded-full bg-emerald-400" />
        <div className="mx-auto flex items-center gap-2 rounded-md border border-slate-200 bg-white px-3 py-1 text-xs text-slate-400">
          <span className="h-2 w-2 rounded-full bg-emerald-400" />
          app.mailgo.io/webmail
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-[180px_1fr]">
        <aside className="hidden flex-col border-r border-slate-200/80 bg-slate-50/40 p-3 sm:flex">
          <button className="mb-4 inline-flex items-center justify-center gap-2 rounded-lg bg-indigo-600 px-3 py-2 text-sm font-semibold text-white">
            Compose
          </button>
          <nav className="space-y-0.5">
            {folders.map((f) => (
              <div
                key={f.name}
                className={
                  'flex items-center justify-between rounded-md px-2.5 py-1.5 text-sm ' +
                  (f.active
                    ? 'bg-indigo-50 font-semibold text-indigo-700'
                    : 'text-slate-600')
                }
              >
                <span>{f.name}</span>
                {f.count ? (
                  <span className="text-xs text-slate-400">{f.count}</span>
                ) : null}
              </div>
            ))}
          </nav>
          <div className="mt-auto rounded-lg border border-slate-200 bg-white p-3">
            <div className="mb-1 flex items-center justify-between text-xs text-slate-500">
              <span>Storage</span>
              <span>6.4 / 10 GB</span>
            </div>
            <div className="h-1.5 w-full overflow-hidden rounded-full bg-slate-100">
              <div className="h-full w-[64%] rounded-full bg-gradient-to-r from-indigo-500 to-blue-500" />
            </div>
          </div>
        </aside>

        <div className="min-w-0">
          <div className="flex items-center justify-between border-b border-slate-200/80 px-4 py-3">
            <div>
              <p className="text-sm font-semibold text-slate-900">Inbox</p>
              <p className="text-xs text-slate-400">24 unread · synced just now</p>
            </div>
            <div className="hidden items-center gap-2 rounded-md border border-slate-200 px-2.5 py-1 text-xs text-slate-400 sm:flex">
              Search mail
            </div>
          </div>

          <ul className="divide-y divide-slate-100">
            {messages.map((m) => (
              <li
                key={m.from}
                className="flex items-center gap-3 px-4 py-3 transition-colors hover:bg-slate-50/70"
              >
                <span className={'flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-semibold text-white ' + m.color}>
                  {m.from.charAt(0)}
                </span>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center justify-between gap-2">
                    <p className={'truncate text-sm ' + (m.unread ? 'font-semibold text-slate-900' : 'text-slate-700')}>
                      {m.from}
                    </p>
                    <span className="shrink-0 text-xs text-slate-400">{m.time}</span>
                  </div>
                  <p className={'truncate text-sm ' + (m.unread ? 'text-slate-700' : 'text-slate-500')}>
                    {m.subject}
                  </p>
                </div>
                {m.unread ? (
                  <span className="h-2 w-2 shrink-0 rounded-full bg-indigo-500" />
                ) : null}
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  )
}
