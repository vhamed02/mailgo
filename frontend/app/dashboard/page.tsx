'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { domainApi, mailboxApi } from '@/lib/api-client'

export default function DashboardPage() {
  const [user, setUser] = useState<any>(null)
  const [org, setOrg] = useState<any>(null)
  const [domainCount, setDomainCount] = useState(0)
  const [mailboxCount, setMailboxCount] = useState(0)

  useEffect(() => {
    const u = localStorage.getItem('user')
    const o = localStorage.getItem('organization')
    if (u) setUser(JSON.parse(u))
    if (o) setOrg(JSON.parse(o))

    domainApi.list().then((r: any) => setDomainCount(r.total ?? 0)).catch(() => {})
    mailboxApi.list().then((r: any) => setMailboxCount(r.total ?? 0)).catch(() => {})
  }, [])

  const stats = [
    { label: 'Total Domains', value: domainCount, icon: 'M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9', bg: 'bg-blue-100', color: 'text-blue-600' },
    { label: 'Active Mailboxes', value: mailboxCount, icon: 'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z', bg: 'bg-indigo-100', color: 'text-indigo-600' },
    { label: 'Storage Used', value: '0 GB', icon: 'M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4', bg: 'bg-purple-100', color: 'text-purple-600' },
    { label: 'Team Members', value: 1, icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z', bg: 'bg-pink-100', color: 'text-pink-600' },
  ]

  const actions = [
    { href: '/dashboard/domains', label: 'Add Domain', sub: 'Connect your email domain', hover: 'hover:border-blue-500 hover:bg-blue-50', iconBg: 'bg-blue-100 group-hover:bg-blue-200', iconColor: 'text-blue-600', activeColor: 'group-hover:text-blue-600' },
    { href: '/dashboard/mailboxes', label: 'Create Mailbox', sub: 'Add a new email account', hover: 'hover:border-indigo-500 hover:bg-indigo-50', iconBg: 'bg-indigo-100 group-hover:bg-indigo-200', iconColor: 'text-indigo-600', activeColor: 'group-hover:text-indigo-600' },
    { href: '/dashboard/settings', label: 'Settings', sub: 'Manage your organisation', hover: 'hover:border-purple-500 hover:bg-purple-50', iconBg: 'bg-purple-100 group-hover:bg-purple-200', iconColor: 'text-purple-600', activeColor: 'group-hover:text-purple-600' },
  ]

  return (
    <div>
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-gray-900 mb-1">
          Welcome back, {user?.first_name}! 👋
        </h2>
        <p className="text-gray-500">Here&apos;s what&apos;s happening with {org?.name} today.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5 mb-8">
        {stats.map((s) => (
          <div key={s.label} className="bg-white p-6 rounded-2xl shadow-sm border border-gray-200">
            <div className={`w-10 h-10 ${s.bg} rounded-xl flex items-center justify-center mb-4`}>
              <svg className={`w-5 h-5 ${s.color}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={s.icon} />
              </svg>
            </div>
            <p className="text-2xl font-bold text-gray-900">{s.value}</p>
            <p className="text-sm text-gray-500 mt-0.5">{s.label}</p>
          </div>
        ))}
      </div>

      <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-6 mb-8">
        <h3 className="text-base font-semibold text-gray-900 mb-4">Quick Actions</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {actions.map((a) => (
            <Link
              key={a.href}
              href={a.href}
              className={`flex items-center p-4 rounded-xl border-2 border-dashed border-gray-300 ${a.hover} transition-all group`}
            >
              <div className={`w-10 h-10 ${a.iconBg} rounded-lg flex items-center justify-center mr-3 transition-colors`}>
                <svg className={`w-5 h-5 ${a.iconColor}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
              </div>
              <div>
                <p className={`font-semibold text-gray-900 ${a.activeColor} transition-colors`}>{a.label}</p>
                <p className="text-sm text-gray-500">{a.sub}</p>
              </div>
            </Link>
          ))}
        </div>
      </div>

      <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-6">
        <h3 className="text-base font-semibold text-gray-900 mb-4">Organisation</h3>
        <dl className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <dt className="text-sm text-gray-500 mb-0.5">Name</dt>
            <dd className="font-semibold text-gray-900">{org?.name}</dd>
          </div>
          <div>
            <dt className="text-sm text-gray-500 mb-0.5">Slug</dt>
            <dd className="font-mono text-sm text-gray-900">{org?.slug}</dd>
          </div>
          <div>
            <dt className="text-sm text-gray-500 mb-0.5">Your Role</dt>
            <dd><span className="px-2.5 py-1 rounded-lg text-xs font-semibold bg-blue-100 text-blue-800">Owner</span></dd>
          </div>
          <div>
            <dt className="text-sm text-gray-500 mb-0.5">Status</dt>
            <dd><span className="inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-semibold bg-green-100 text-green-800"><span className="w-1.5 h-1.5 bg-green-500 rounded-full mr-1.5" />Active</span></dd>
          </div>
        </dl>
      </div>
    </div>
  )
}
