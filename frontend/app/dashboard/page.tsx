'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { ComingSoonModal, type ComingSoonFeature } from '@/components/ui/coming-soon-modal'

/* ── "Coming soon" feature definitions ───────────────────────────────── */
const FEATURES: Record<string, ComingSoonFeature> = {
  domain: {
    title: 'Domain Management',
    description: 'Connect and manage your custom email domains in one place.',
    iconBg: 'bg-blue-100',
    iconColor: 'text-blue-600',
    iconPath: (
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"
      />
    ),
    items: [
      'Add custom domains (e.g. company.com)',
      'Verify DNS records automatically',
      'Configure SPF, DKIM, and DMARC',
      'Manage domain settings',
    ],
  },
  mailbox: {
    title: 'Mailbox Management',
    description: 'Create and manage email accounts for your entire team.',
    iconBg: 'bg-indigo-100',
    iconColor: 'text-indigo-600',
    iconPath: (
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
      />
    ),
    items: [
      'Create email accounts',
      'Set per-mailbox storage quotas',
      'Manage passwords securely',
      'Suspend or unsuspend accounts',
    ],
  },
  team: {
    title: 'Team Invitations',
    description: 'Invite colleagues and manage their access to your organisation.',
    iconBg: 'bg-purple-100',
    iconColor: 'text-purple-600',
    iconPath: (
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
      />
    ),
    items: [
      'Invite team members by email',
      'Assign roles (admin / user)',
      'Manage permissions per role',
      'Track pending invitations',
    ],
  },
  setup: {
    title: 'Setup Guide',
    description: 'A step-by-step walkthrough to get your email infrastructure running.',
    iconBg: 'bg-green-100',
    iconColor: 'text-green-600',
    iconPath: (
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z"
      />
    ),
    items: [
      'Add your company domain',
      'Verify DNS with SPF, DKIM, and DMARC',
      'Create mailboxes for your team',
      'Send a test email to confirm everything works',
    ],
  },
}

/* ── Page ────────────────────────────────────────────────────────────── */
export default function DashboardPage() {
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [organization, setOrganization] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [activeModal, setActiveModal] = useState<keyof typeof FEATURES | null>(null)

  useEffect(() => {
    const token = localStorage.getItem('access_token')
    if (!token) {
      router.push('/auth/login')
      return
    }
    const userData = localStorage.getItem('user')
    const orgData = localStorage.getItem('organization')
    if (userData) setUser(JSON.parse(userData))
    if (orgData) setOrganization(JSON.parse(orgData))
    setLoading(false)
  }, [router])

  const handleLogout = () => {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user')
    localStorage.removeItem('organization')
    router.push('/auth/login')
  }

  const openModal = (key: keyof typeof FEATURES) => setActiveModal(key)
  const closeModal = () => setActiveModal(null)

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600" />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
      {/* ── Header ──────────────────────────────────────────────────── */}
      <header className="bg-white border-b border-gray-200 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <div className="w-10 h-10 bg-gradient-to-br from-blue-600 to-indigo-600 rounded-xl flex items-center justify-center mr-3">
                <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
              </div>
              <h1 className="text-xl font-bold text-gray-900">MailGo</h1>
            </div>
            <Button onClick={handleLogout} variant="outline" className="rounded-lg">
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
              </svg>
              Sign out
            </Button>
          </div>
        </div>
      </header>

      {/* ── Main Content ─────────────────────────────────────────────── */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Welcome */}
        <div className="mb-8">
          <h2 className="text-3xl font-bold text-gray-900 mb-2">
            Welcome back, {user?.first_name}! 👋
          </h2>
          <p className="text-gray-600">
            Here&apos;s what&apos;s happening with {organization?.name} today.
          </p>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-200 hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
                <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                </svg>
              </div>
              <span className="text-sm font-medium text-green-600 bg-green-100 px-2 py-1 rounded-lg">+0%</span>
            </div>
            <h3 className="text-2xl font-bold text-gray-900 mb-1">0</h3>
            <p className="text-sm text-gray-600">Total Domains</p>
          </div>

          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-200 hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 bg-indigo-100 rounded-xl flex items-center justify-center">
                <svg className="w-6 h-6 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
              </div>
              <span className="text-sm font-medium text-green-600 bg-green-100 px-2 py-1 rounded-lg">+0%</span>
            </div>
            <h3 className="text-2xl font-bold text-gray-900 mb-1">0</h3>
            <p className="text-sm text-gray-600">Active Mailboxes</p>
          </div>

          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-200 hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center">
                <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
                </svg>
              </div>
              <span className="text-sm font-medium text-gray-500 bg-gray-100 px-2 py-1 rounded-lg">0%</span>
            </div>
            <h3 className="text-2xl font-bold text-gray-900 mb-1">0 GB</h3>
            <p className="text-sm text-gray-600">of 10 GB used</p>
          </div>

          <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-200 hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 bg-pink-100 rounded-xl flex items-center justify-center">
                <svg className="w-6 h-6 text-pink-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                </svg>
              </div>
              <span className="text-sm font-medium text-gray-500 bg-gray-100 px-2 py-1 rounded-lg">+100%</span>
            </div>
            <h3 className="text-2xl font-bold text-gray-900 mb-1">1</h3>
            <p className="text-sm text-gray-600">Team Members</p>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-200 mb-8">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <button
              onClick={() => openModal('domain')}
              className="flex items-center p-4 rounded-xl border-2 border-dashed border-gray-300 hover:border-blue-500 hover:bg-blue-50 transition-all group cursor-pointer"
            >
              <div className="w-10 h-10 bg-blue-100 group-hover:bg-blue-200 rounded-lg flex items-center justify-center mr-3 transition-colors">
                <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
              </div>
              <div className="text-left">
                <p className="font-semibold text-gray-900 group-hover:text-blue-600 transition-colors">Add Domain</p>
                <p className="text-sm text-gray-600">Connect your email domain</p>
              </div>
            </button>

            <button
              onClick={() => openModal('mailbox')}
              className="flex items-center p-4 rounded-xl border-2 border-dashed border-gray-300 hover:border-indigo-500 hover:bg-indigo-50 transition-all group cursor-pointer"
            >
              <div className="w-10 h-10 bg-indigo-100 group-hover:bg-indigo-200 rounded-lg flex items-center justify-center mr-3 transition-colors">
                <svg className="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
              </div>
              <div className="text-left">
                <p className="font-semibold text-gray-900 group-hover:text-indigo-600 transition-colors">Create Mailbox</p>
                <p className="text-sm text-gray-600">Add a new email account</p>
              </div>
            </button>

            <button
              onClick={() => openModal('team')}
              className="flex items-center p-4 rounded-xl border-2 border-dashed border-gray-300 hover:border-purple-500 hover:bg-purple-50 transition-all group cursor-pointer"
            >
              <div className="w-10 h-10 bg-purple-100 group-hover:bg-purple-200 rounded-lg flex items-center justify-center mr-3 transition-colors">
                <svg className="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                </svg>
              </div>
              <div className="text-left">
                <p className="font-semibold text-gray-900 group-hover:text-purple-600 transition-colors">Invite Team</p>
                <p className="text-sm text-gray-600">Add team members</p>
              </div>
            </button>
          </div>
        </div>

        {/* Getting Started */}
        <div className="bg-gradient-to-br from-blue-600 to-indigo-700 p-8 rounded-2xl shadow-lg text-white">
          <h3 className="text-2xl font-bold mb-2">🚀 Getting Started with MailGo</h3>
          <p className="text-blue-100 mb-6">Follow these steps to set up your professional email</p>
          <div className="space-y-4">
            <div className="flex items-start">
              <div className="w-8 h-8 bg-white/20 backdrop-blur rounded-lg flex items-center justify-center mr-4 flex-shrink-0 font-bold">1</div>
              <div>
                <h4 className="font-semibold mb-1">Add your domain</h4>
                <p className="text-blue-100 text-sm">Connect your company domain (e.g., company.com)</p>
              </div>
            </div>
            <div className="flex items-start">
              <div className="w-8 h-8 bg-white/20 backdrop-blur rounded-lg flex items-center justify-center mr-4 flex-shrink-0 font-bold">2</div>
              <div>
                <h4 className="font-semibold mb-1">Verify DNS records</h4>
                <p className="text-blue-100 text-sm">Add SPF, DKIM, and DMARC records to your DNS</p>
              </div>
            </div>
            <div className="flex items-start">
              <div className="w-8 h-8 bg-white/20 backdrop-blur rounded-lg flex items-center justify-center mr-4 flex-shrink-0 font-bold">3</div>
              <div>
                <h4 className="font-semibold mb-1">Create mailboxes</h4>
                <p className="text-blue-100 text-sm">Set up email accounts for your team</p>
              </div>
            </div>
          </div>
          <Button
            onClick={() => openModal('setup')}
            className="mt-6 bg-white text-blue-600 hover:bg-blue-50 rounded-lg font-semibold"
          >
            Start Setup Guide
          </Button>
        </div>

        {/* Organization Info */}
        <div className="mt-8 bg-white p-6 rounded-2xl shadow-sm border border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Organization Information</h3>
          <dl className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <dt className="text-sm font-medium text-gray-600 mb-1">Organization Name</dt>
              <dd className="text-base font-semibold text-gray-900">{organization?.name}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-600 mb-1">Organization Slug</dt>
              <dd className="text-base font-mono text-gray-900">{organization?.slug}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-600 mb-1">Your Role</dt>
              <dd>
                <span className="inline-flex items-center px-3 py-1 rounded-lg text-sm font-semibold bg-blue-100 text-blue-800">
                  Owner
                </span>
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-600 mb-1">Account Status</dt>
              <dd>
                <span className="inline-flex items-center px-3 py-1 rounded-lg text-sm font-semibold bg-green-100 text-green-800">
                  <span className="w-2 h-2 bg-green-500 rounded-full mr-2" />
                  Active
                </span>
              </dd>
            </div>
          </dl>
        </div>
      </main>

      {/* ── Modals (single mount, driven by activeModal state) ────────── */}
      {activeModal && (
        <ComingSoonModal
          open={activeModal !== null}
          onOpenChange={(open) => { if (!open) closeModal() }}
          feature={FEATURES[activeModal]}
        />
      )}
    </div>
  )
}
