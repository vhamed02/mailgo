import { Container } from '../container'

const STATS = [
  { value: '99.9%', label: 'Uptime SLA' },
  { value: '128K+', label: 'Mailboxes hosted' },
  { value: '10 GB', label: 'Storage per user' },
  { value: '24/7', label: 'Expert support' },
]

export function Stats() {
  return (
    <section className="bg-white py-16 sm:py-20">
      <Container>
        <div className="overflow-hidden rounded-3xl border border-slate-200 bg-gradient-to-br from-indigo-50 via-white to-blue-50 px-6 py-10 sm:px-12 sm:py-12">
          <div className="grid grid-cols-2 gap-y-10 lg:grid-cols-4 lg:divide-x lg:divide-slate-200">
            {STATS.map((stat) => (
              <div key={stat.label} className="text-center lg:px-6">
                <p className="text-4xl font-bold tracking-tight text-slate-900 sm:text-5xl">
                  {stat.value}
                </p>
                <p className="mt-2 text-sm font-medium text-slate-500">
                  {stat.label}
                </p>
              </div>
            ))}
          </div>
        </div>
      </Container>
    </section>
  )
}
