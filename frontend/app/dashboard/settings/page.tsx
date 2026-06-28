'use client'

import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export default function SettingsPage() {
  const [user, setUser] = useState<any>(null)
  const [org, setOrg] = useState<any>(null)
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    const u = localStorage.getItem('user')
    const o = localStorage.getItem('organization')
    if (u) setUser(JSON.parse(u))
    if (o) setOrg(JSON.parse(o))
  }, [])

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault()
    setSaved(true)
    setTimeout(() => setSaved(false), 3000)
  }

  return (
    <div className="max-w-2xl">
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-gray-900">Settings</h2>
        <p className="text-gray-500 mt-0.5">Manage your account and organisation</p>
      </div>

      <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-6 mb-6">
        <h3 className="text-base font-semibold text-gray-900 mb-4">Profile</h3>
        <div className="flex items-center gap-4 mb-6">
          <div className="w-14 h-14 bg-gradient-to-br from-blue-500 to-indigo-500 rounded-full flex items-center justify-center text-white text-lg font-bold">
            {user?.first_name?.[0]}{user?.last_name?.[0]}
          </div>
          <div>
            <p className="font-semibold text-gray-900">{user?.first_name} {user?.last_name}</p>
            <p className="text-sm text-gray-500">{user?.email}</p>
          </div>
        </div>
        <dl className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <dt className="text-gray-500 mb-0.5">Role</dt>
            <dd><span className="px-2.5 py-1 rounded-lg text-xs font-semibold bg-blue-100 text-blue-800">Owner</span></dd>
          </div>
          <div>
            <dt className="text-gray-500 mb-0.5">Status</dt>
            <dd><span className="inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-semibold bg-green-100 text-green-800"><span className="w-1.5 h-1.5 bg-green-500 rounded-full mr-1.5" />Active</span></dd>
          </div>
        </dl>
      </div>

      <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-6 mb-6">
        <h3 className="text-base font-semibold text-gray-900 mb-4">Organisation</h3>
        <form onSubmit={handleSave} className="space-y-4">
          <div>
            <Label className="mb-2 block">Organisation name</Label>
            <Input
              defaultValue={org?.name}
              onChange={(e) => setOrg({ ...org, name: e.target.value })}
            />
          </div>
          <div>
            <Label className="mb-2 block">Slug</Label>
            <Input
              defaultValue={org?.slug}
              onChange={(e) => setOrg({ ...org, slug: e.target.value })}
            />
            <p className="text-xs text-gray-500 mt-1">Lowercase letters, numbers, and hyphens only.</p>
          </div>
          <div className="flex items-center gap-3 pt-2">
            <Button type="submit" className="rounded-lg">
              {saved ? '✓ Saved' : 'Save Changes'}
            </Button>
            {saved && <p className="text-sm text-green-600">Changes saved successfully.</p>}
          </div>
        </form>
      </div>

      <div className="bg-white rounded-2xl shadow-sm border border-red-200 p-6">
        <h3 className="text-base font-semibold text-red-700 mb-2">Danger Zone</h3>
        <p className="text-sm text-gray-600 mb-4">Permanently delete your organisation and all associated data. This cannot be undone.</p>
        <Button variant="outline" className="rounded-lg border-red-300 text-red-600 hover:bg-red-50" disabled>
          Delete Organisation
        </Button>
        <p className="text-xs text-gray-400 mt-2">Contact support to delete your organisation.</p>
      </div>
    </div>
  )
}
