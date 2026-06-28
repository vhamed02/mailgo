'use client'

import { useEffect, useState } from 'react'
import { domainApi } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Modal, ModalContent, ModalHeader, ModalTitle, ModalDescription, ModalFooter, ModalClose } from '@/components/ui/modal'

type Domain = {
  id: string
  name: string
  status: string
  dns_verified: boolean
  spf_record: string
  dkim_record: string
  dmarc_record: string
  created_at: string
}

function StatusBadge({ verified, status }: { verified: boolean; status: string }) {
  if (verified) return <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800"><span className="w-1.5 h-1.5 rounded-full bg-green-500" />Active</span>
  if (status === 'pending') return <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800"><span className="w-1.5 h-1.5 rounded-full bg-yellow-400" />Pending</span>
  return <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-600"><span className="w-1.5 h-1.5 rounded-full bg-gray-400" />{status}</span>
}

export default function DomainsPage() {
  const [domains, setDomains] = useState<Domain[]>([])
  const [loading, setLoading] = useState(true)
  const [addOpen, setAddOpen] = useState(false)
  const [dnsOpen, setDnsOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [selected, setSelected] = useState<Domain | null>(null)
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const [verifying, setVerifying] = useState<string | null>(null)
  const [copied, setCopied] = useState<string | null>(null)

  const load = () => {
    setLoading(true)
    domainApi.list().then((r: any) => setDomains(r.data ?? [])).finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSaving(true)
    try {
      await domainApi.create({ name })
      setName('')
      setAddOpen(false)
      load()
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to add domain')
    } finally {
      setSaving(false)
    }
  }

  const handleVerify = async (id: string) => {
    setVerifying(id)
    try {
      await domainApi.verify(id)
      load()
    } finally {
      setVerifying(null)
    }
  }

  const handleDelete = async () => {
    if (!selected) return
    try {
      await domainApi.delete(selected.id)
      setDeleteOpen(false)
      setSelected(null)
      load()
    } catch {}
  }

  const copy = (text: string, key: string) => {
    navigator.clipboard.writeText(text)
    setCopied(key)
    setTimeout(() => setCopied(null), 2000)
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Domains</h2>
          <p className="text-gray-500 mt-0.5">Manage your email domains</p>
        </div>
        <Button onClick={() => setAddOpen(true)} className="rounded-xl">
          <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Domain
        </Button>
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
        </div>
      ) : domains.length === 0 ? (
        <div className="bg-white rounded-2xl border-2 border-dashed border-gray-300 p-16 text-center">
          <div className="w-16 h-16 bg-blue-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-1">No domains yet</h3>
          <p className="text-gray-500 mb-6">Add your first domain to start creating mailboxes.</p>
          <Button onClick={() => setAddOpen(true)} className="rounded-xl">Add your first domain</Button>
        </div>
      ) : (
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100 bg-gray-50">
                <th className="text-left text-xs font-semibold text-gray-500 uppercase tracking-wide px-6 py-3">Domain</th>
                <th className="text-left text-xs font-semibold text-gray-500 uppercase tracking-wide px-6 py-3">Status</th>
                <th className="text-left text-xs font-semibold text-gray-500 uppercase tracking-wide px-6 py-3">Added</th>
                <th className="px-6 py-3" />
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {domains.map((d) => (
                <tr key={d.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4 font-medium text-gray-900">{d.name}</td>
                  <td className="px-6 py-4"><StatusBadge verified={d.dns_verified} status={d.status} /></td>
                  <td className="px-6 py-4 text-sm text-gray-500">{new Date(d.created_at).toLocaleDateString()}</td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2 justify-end">
                      <Button
                        variant="outline"
                        size="sm"
                        className="rounded-lg text-xs"
                        onClick={() => { setSelected(d); setDnsOpen(true) }}
                      >
                        DNS Records
                      </Button>
                      {!d.dns_verified && (
                        <Button
                          size="sm"
                          className="rounded-lg text-xs"
                          onClick={() => handleVerify(d.id)}
                          disabled={verifying === d.id}
                        >
                          {verifying === d.id ? 'Checking...' : 'Verify'}
                        </Button>
                      )}
                      <button
                        onClick={() => { setSelected(d); setDeleteOpen(true) }}
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

      <Modal open={addOpen} onOpenChange={setAddOpen}>
        <ModalContent>
          <ModalHeader>
            <ModalTitle>Add Domain</ModalTitle>
            <ModalDescription>Enter the domain you want to use for email hosting.</ModalDescription>
          </ModalHeader>
          <form onSubmit={handleAdd}>
            <div className="mb-4">
              <Label htmlFor="domain-name" className="mb-2 block">Domain name</Label>
              <Input
                id="domain-name"
                placeholder="company.com"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
              {error && <p className="text-sm text-red-600 mt-2">{error}</p>}
            </div>
            <ModalFooter>
              <ModalClose asChild>
                <Button type="button" variant="outline" className="rounded-lg">Cancel</Button>
              </ModalClose>
              <Button type="submit" className="rounded-lg" disabled={saving}>
                {saving ? 'Adding...' : 'Add Domain'}
              </Button>
            </ModalFooter>
          </form>
        </ModalContent>
      </Modal>

      <Modal open={dnsOpen} onOpenChange={setDnsOpen}>
        <ModalContent className="max-w-2xl">
          <ModalHeader>
            <ModalTitle>DNS Records — {selected?.name}</ModalTitle>
            <ModalDescription>Add these records to your DNS provider to verify ownership and enable email.</ModalDescription>
          </ModalHeader>
          <div className="space-y-4 my-2">
            {selected && [
              { type: 'TXT (SPF)', key: 'spf', value: selected.spf_record },
              { type: 'TXT (DKIM)', key: 'dkim', value: selected.dkim_record },
              { type: 'TXT (DMARC)', key: 'dmarc', value: selected.dmarc_record },
            ].map((rec) => (
              <div key={rec.key} className="bg-gray-50 rounded-xl p-4">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs font-semibold text-gray-500 uppercase tracking-wide">{rec.type}</span>
                  <button
                    onClick={() => copy(rec.value, rec.key)}
                    className="text-xs text-blue-600 hover:text-blue-700 font-medium"
                  >
                    {copied === rec.key ? 'Copied!' : 'Copy'}
                  </button>
                </div>
                <p className="text-sm font-mono text-gray-800 break-all">{rec.value}</p>
              </div>
            ))}
          </div>
          <ModalFooter>
            <ModalClose asChild>
              <Button variant="outline" className="rounded-lg">Close</Button>
            </ModalClose>
            {selected && !selected.dns_verified && (
              <Button
                className="rounded-lg"
                onClick={() => { handleVerify(selected.id); setDnsOpen(false) }}
              >
                Verify Now
              </Button>
            )}
          </ModalFooter>
        </ModalContent>
      </Modal>

      <Modal open={deleteOpen} onOpenChange={setDeleteOpen}>
        <ModalContent>
          <ModalHeader>
            <ModalTitle>Delete Domain</ModalTitle>
            <ModalDescription>
              Are you sure you want to delete <strong>{selected?.name}</strong>? This will also delete all mailboxes under this domain.
            </ModalDescription>
          </ModalHeader>
          <ModalFooter>
            <ModalClose asChild>
              <Button variant="outline" className="rounded-lg">Cancel</Button>
            </ModalClose>
            <Button
              className="rounded-lg bg-red-600 hover:bg-red-700 text-white"
              onClick={handleDelete}
            >
              Delete
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </div>
  )
}
