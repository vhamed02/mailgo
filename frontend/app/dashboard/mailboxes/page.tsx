'use client'

import { useEffect, useState } from 'react'
import { mailboxApi, domainApi } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Modal, ModalContent, ModalHeader, ModalTitle, ModalDescription, ModalFooter, ModalClose } from '@/components/ui/modal'

type Mailbox = {
  id: string
  email: string
  display_name: string
  status: string
  quota_bytes: number
  used_bytes: number
  domain_id: string
  created_at: string
}

type Domain = { id: string; name: string; dns_verified: boolean }

function StatusBadge({ status }: { status: string }) {
  if (status === 'active') return <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800"><span className="w-1.5 h-1.5 rounded-full bg-green-500" />Active</span>
  if (status === 'suspended') return <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800"><span className="w-1.5 h-1.5 rounded-full bg-yellow-400" />Suspended</span>
  return <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-600">{status}</span>
}

const fmt = (bytes: number) => {
  if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(1)} GB`
  if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(0)} MB`
  return `${bytes} B`
}

export default function MailboxesPage() {
  const [mailboxes, setMailboxes] = useState<Mailbox[]>([])
  const [domains, setDomains] = useState<Domain[]>([])
  const [loading, setLoading] = useState(true)
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [selected, setSelected] = useState<Mailbox | null>(null)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  const [form, setForm] = useState({ domain_id: '', local_part: '', password: '', display_name: '', quota_gb: '1' })
  const [editForm, setEditForm] = useState({ display_name: '', password: '' })

  const openWebmail = (email: string, id: string) => {
    const allEmails = mailboxes.map(m => m.email)
    sessionStorage.setItem('webmail_mailboxes', JSON.stringify(allEmails))
    window.open(`/webmail?mailbox=${encodeURIComponent(email)}&id=${encodeURIComponent(id)}`, '_blank')
  }

  const load = () => {
    setLoading(true)
    Promise.all([
      mailboxApi.list().then((r: any) => setMailboxes(r.data ?? [])),
      domainApi.list().then((r: any) => setDomains((r.data ?? []).filter((d: Domain) => d.dns_verified))),
    ]).finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSaving(true)
    try {
      await mailboxApi.create({
        domain_id: form.domain_id,
        local_part: form.local_part,
        password: form.password,
        display_name: form.display_name,
        quota_bytes: parseInt(form.quota_gb) * 1073741824,
      })
      setCreateOpen(false)
      setForm({ domain_id: '', local_part: '', password: '', display_name: '', quota_gb: '1' })
      load()
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to create mailbox')
    } finally {
      setSaving(false)
    }
  }

  const handleEdit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selected) return
    setSaving(true)
    try {
      const payload: any = {}
      if (editForm.display_name) payload.display_name = editForm.display_name
      if (editForm.password) payload.password = editForm.password
      await mailboxApi.update(selected.id, payload)
      setEditOpen(false)
      load()
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to update mailbox')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!selected) return
    try {
      await mailboxApi.delete(selected.id)
      setDeleteOpen(false)
      setSelected(null)
      load()
    } catch {}
  }

  const handleSuspend = async (m: Mailbox) => {
    try {
      if (m.status === 'active') await mailboxApi.suspend(m.id)
      else await mailboxApi.unsuspend(m.id)
      load()
    } catch {}
  }

  const verifiedDomains = domains.filter((d) => d.dns_verified)

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Mailboxes</h2>
          <p className="text-gray-500 mt-0.5">Manage email accounts for your team</p>
        </div>
        <Button
          onClick={() => { setError(''); setCreateOpen(true) }}
          className="rounded-xl"
          disabled={verifiedDomains.length === 0}
          title={verifiedDomains.length === 0 ? 'You need a verified domain first' : ''}
        >
          <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Create Mailbox
        </Button>
      </div>

      {verifiedDomains.length === 0 && !loading && (
        <div className="mb-6 p-4 bg-amber-50 border border-amber-200 rounded-xl flex items-start gap-3">
          <svg className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <div className="text-sm text-amber-800">
            You need a <strong>verified domain</strong> before creating mailboxes.{' '}
            <a href="/dashboard/domains" className="underline font-medium">Add a domain →</a>
          </div>
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
        </div>
      ) : mailboxes.length === 0 ? (
        <div className="bg-white rounded-2xl border-2 border-dashed border-gray-300 p-16 text-center">
          <div className="w-16 h-16 bg-indigo-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-1">No mailboxes yet</h3>
          <p className="text-gray-500 mb-6">Create your first mailbox after verifying a domain.</p>
        </div>
      ) : (
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100 bg-gray-50">
                <th className="text-left text-xs font-semibold text-gray-500 uppercase tracking-wide px-6 py-3">Email</th>
                <th className="text-left text-xs font-semibold text-gray-500 uppercase tracking-wide px-6 py-3">Display Name</th>
                <th className="text-left text-xs font-semibold text-gray-500 uppercase tracking-wide px-6 py-3">Status</th>
                <th className="text-left text-xs font-semibold text-gray-500 uppercase tracking-wide px-6 py-3">Quota</th>
                <th className="px-6 py-3" />
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {mailboxes.map((m) => (
                <tr key={m.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4 font-medium text-gray-900">{m.email}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">{m.display_name || '—'}</td>
                  <td className="px-6 py-4"><StatusBadge status={m.status} /></td>
                  <td className="px-6 py-4 text-sm text-gray-600">{fmt(m.used_bytes)} / {fmt(m.quota_bytes)}</td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2 justify-end">
                        <button
                          onClick={() => openWebmail(m.email, m.id)}
                          className="inline-flex items-center justify-center h-9 px-3 rounded-lg text-xs font-medium border border-indigo-300 text-indigo-600 hover:bg-indigo-50 transition-colors"
                        >
                        <svg className="w-3.5 h-3.5 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                        </svg>
                        Webmail
                      </button>
                      <Button
                        variant="outline"
                        size="sm"
                        className="rounded-lg text-xs"
                        onClick={() => { setSelected(m); setEditForm({ display_name: m.display_name, password: '' }); setError(''); setEditOpen(true) }}
                      >
                        Edit
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        className={`rounded-lg text-xs ${m.status === 'active' ? 'text-yellow-600 border-yellow-300 hover:bg-yellow-50' : 'text-green-600 border-green-300 hover:bg-green-50'}`}
                        onClick={() => handleSuspend(m)}
                      >
                        {m.status === 'active' ? 'Suspend' : 'Unsuspend'}
                      </Button>
                      <button
                        onClick={() => { setSelected(m); setDeleteOpen(true) }}
                        className="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition-colors"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Modal open={createOpen} onOpenChange={setCreateOpen}>
        <ModalContent>
          <ModalHeader>
            <ModalTitle>Create Mailbox</ModalTitle>
            <ModalDescription>Set up a new email account on one of your verified domains.</ModalDescription>
          </ModalHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <Label className="mb-2 block">Domain</Label>
              <select
                className="w-full h-10 rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                value={form.domain_id}
                onChange={(e) => setForm({ ...form, domain_id: e.target.value })}
                required
              >
                <option value="">Select a domain</option>
                {verifiedDomains.map((d) => (
                  <option key={d.id} value={d.id}>{d.name}</option>
                ))}
              </select>
            </div>
            <div>
              <Label className="mb-2 block">Local part</Label>
              <div className="flex items-center gap-2">
                <Input
                  placeholder="info"
                  value={form.local_part}
                  onChange={(e) => setForm({ ...form, local_part: e.target.value })}
                  required
                />
                <span className="text-gray-500 text-sm flex-shrink-0">
                  @{verifiedDomains.find((d) => d.id === form.domain_id)?.name ?? 'domain.com'}
                </span>
              </div>
            </div>
            <div>
              <Label className="mb-2 block">Display name</Label>
              <Input
                placeholder="Info Team"
                value={form.display_name}
                onChange={(e) => setForm({ ...form, display_name: e.target.value })}
                required
              />
            </div>
            <div>
              <Label className="mb-2 block">Password</Label>
              <Input
                type="password"
                placeholder="At least 8 characters"
                value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })}
                required
                minLength={8}
              />
            </div>
            <div>
              <Label className="mb-2 block">Quota (GB)</Label>
              <Input
                type="number"
                min="1"
                max="50"
                value={form.quota_gb}
                onChange={(e) => setForm({ ...form, quota_gb: e.target.value })}
                required
              />
            </div>
            {error && <p className="text-sm text-red-600">{error}</p>}
            <ModalFooter>
              <ModalClose asChild>
                <Button type="button" variant="outline" className="rounded-lg">Cancel</Button>
              </ModalClose>
              <Button type="submit" className="rounded-lg" disabled={saving}>
                {saving ? 'Creating...' : 'Create Mailbox'}
              </Button>
            </ModalFooter>
          </form>
        </ModalContent>
      </Modal>

      <Modal open={editOpen} onOpenChange={setEditOpen}>
        <ModalContent>
          <ModalHeader>
            <ModalTitle>Edit Mailbox</ModalTitle>
            <ModalDescription>{selected?.email}</ModalDescription>
          </ModalHeader>
          <form onSubmit={handleEdit} className="space-y-4">
            <div>
              <Label className="mb-2 block">Display name</Label>
              <Input
                value={editForm.display_name}
                onChange={(e) => setEditForm({ ...editForm, display_name: e.target.value })}
              />
            </div>
            <div>
              <Label className="mb-2 block">New password <span className="text-gray-400 font-normal">(leave blank to keep current)</span></Label>
              <Input
                type="password"
                placeholder="New password"
                value={editForm.password}
                onChange={(e) => setEditForm({ ...editForm, password: e.target.value })}
                minLength={8}
              />
            </div>
            {error && <p className="text-sm text-red-600">{error}</p>}
            <ModalFooter>
              <ModalClose asChild>
                <Button type="button" variant="outline" className="rounded-lg">Cancel</Button>
              </ModalClose>
              <Button type="submit" className="rounded-lg" disabled={saving}>
                {saving ? 'Saving...' : 'Save Changes'}
              </Button>
            </ModalFooter>
          </form>
        </ModalContent>
      </Modal>

      <Modal open={deleteOpen} onOpenChange={setDeleteOpen}>
        <ModalContent>
          <ModalHeader>
            <ModalTitle>Delete Mailbox</ModalTitle>
            <ModalDescription>
              Are you sure you want to delete <strong>{selected?.email}</strong>? This action cannot be undone.
            </ModalDescription>
          </ModalHeader>
          <ModalFooter>
            <ModalClose asChild>
              <Button variant="outline" className="rounded-lg">Cancel</Button>
            </ModalClose>
            <Button className="rounded-lg bg-red-600 hover:bg-red-700 text-white" onClick={handleDelete}>
              Delete
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </div>
  )
}
