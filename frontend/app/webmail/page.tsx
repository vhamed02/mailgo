'use client'

import { useEffect, useState, useCallback, Suspense } from 'react'
import { useSearchParams } from 'next/navigation'
import { createMailClient, type MailFolder, type MailMessage, type ComposePayload } from '@/lib/mail-api'

function formatDate(d: string) {
  const date = new Date(d)
  const now = new Date()
  if (date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
}

function senderName(from: string) {
  const m = from.match(/^(.+?)\s*</)
  return m ? m[1].trim().replace(/^"(.*)"$/, '$1') : from.split('@')[0]
}

function UnlockScreen({ mailbox, onUnlock }: { mailbox: string; onUnlock: (p: string) => void }) {
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const client = createMailClient(mailbox, password)
      await client.listFolders()
      onUnlock(password)
    } catch {
      setError('Invalid password or mailbox unreachable.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-xl p-8 w-full max-w-sm">
        <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center mx-auto mb-4">
          <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
          </svg>
        </div>
        <h1 className="text-xl font-bold text-gray-900 text-center mb-1">Open Mailbox</h1>
        <p className="text-sm text-gray-500 text-center mb-6 font-mono">{mailbox}</p>
        <form onSubmit={handleSubmit} className="space-y-4">
          <input
            type="password"
            placeholder="Mailbox password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            required
            className="w-full h-11 px-4 rounded-xl border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          {error && <p className="text-sm text-red-600">{error}</p>}
          <button
            type="submit"
            disabled={loading}
            className="w-full h-11 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-semibold text-sm transition-colors disabled:opacity-50"
          >
            {loading ? 'Connecting...' : 'Open Webmail'}
          </button>
        </form>
      </div>
    </div>
  )
}

function ComposeModal({ from, onClose, onSend, replyTo }: {
  from: string
  onClose: () => void
  onSend: (p: ComposePayload) => Promise<void>
  replyTo?: MailMessage
}) {
  const [to, setTo] = useState(replyTo ? replyTo.from : '')
  const [subject, setSubject] = useState(replyTo ? (replyTo.subject.startsWith('Re:') ? replyTo.subject : `Re: ${replyTo.subject}`) : '')
  const [body, setBody] = useState('')
  const [sending, setSending] = useState(false)
  const [error, setError] = useState('')

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault()
    setSending(true)
    setError('')
    try {
      await onSend({
        to: to.split(',').map(s => s.trim()).filter(Boolean),
        subject,
        body,
        is_html: false,
        in_reply_to: replyTo?.message_id,
      })
      onClose()
    } catch {
      setError('Failed to send. Check your connection.')
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-end p-4 pointer-events-none">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-lg pointer-events-auto" style={{ maxHeight: '80vh' }}>
        <div className="flex items-center justify-between px-5 py-3 border-b border-gray-100 bg-gray-50 rounded-t-2xl">
          <span className="font-semibold text-sm text-gray-800">{replyTo ? 'Reply' : 'New Message'}</span>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <form onSubmit={handleSend} className="flex flex-col p-4 gap-3">
          <input placeholder="To" value={to} onChange={e => setTo(e.target.value)} required
            className="w-full h-9 px-3 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <input placeholder="Subject" value={subject} onChange={e => setSubject(e.target.value)} required
            className="w-full h-9 px-3 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <textarea placeholder="Write your message..." value={body} onChange={e => setBody(e.target.value)} required rows={8}
            className="w-full px-3 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none" />
          {error && <p className="text-xs text-red-600">{error}</p>}
          <div className="flex justify-end">
            <button type="submit" disabled={sending}
              className="h-9 px-5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-semibold text-sm transition-colors disabled:opacity-50">
              {sending ? 'Sending...' : 'Send'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

function WebmailApp({ mailboxes, initialMailbox, password }: {
  mailboxes: string[]
  initialMailbox: string
  password: string
}) {
  const [mailbox, setMailbox] = useState(initialMailbox)
  const [client, setClient] = useState(() => createMailClient(initialMailbox, password))
  const [folders, setFolders] = useState<MailFolder[]>([])
  const [activeFolder, setActiveFolder] = useState('INBOX')
  const [messages, setMessages] = useState<MailMessage[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [selected, setSelected] = useState<MailMessage | null>(null)
  const [full, setFull] = useState<MailMessage | null>(null)
  const [loadingMsgs, setLoadingMsgs] = useState(false)
  const [compose, setCompose] = useState(false)
  const [replyMsg, setReplyMsg] = useState<MailMessage | undefined>()

  const switchMailbox = (addr: string) => {
    setMailbox(addr)
    setClient(createMailClient(addr, password))
    setActiveFolder('INBOX')
    setSelected(null)
    setFull(null)
    setPage(1)
  }

  const loadFolders = useCallback(async () => {
    try {
      const f = await client.listFolders()
      setFolders(f)
    } catch {}
  }, [client])

  const loadMessages = useCallback(async () => {
    setLoadingMsgs(true)
    try {
      const r = await client.listMessages(activeFolder, page)
      setMessages(r.data)
      setTotal(r.total)
    } catch {} finally {
      setLoadingMsgs(false)
    }
  }, [client, activeFolder, page])

  useEffect(() => { loadFolders() }, [loadFolders])
  useEffect(() => { loadMessages(); setSelected(null); setFull(null) }, [loadMessages])

  const openMessage = async (msg: MailMessage) => {
    setSelected(msg)
    try {
      const m = await client.getMessage(msg.uid, activeFolder)
      setFull(m)
    } catch {}
  }

  const deleteMessage = async (msg: MailMessage) => {
    try {
      await client.deleteMessage(msg.uid, activeFolder)
      setSelected(null); setFull(null)
      loadMessages()
    } catch {}
  }

  const sendMessage = async (payload: ComposePayload) => {
    if (replyMsg) {
      await client.reply(payload)
    } else {
      await client.compose(payload)
    }
    loadMessages()
  }

  const folderOrder = ['INBOX', 'Sent', 'Drafts', 'Spam', 'Trash']
  const sortedFolders = [
    ...folderOrder.map(n => folders.find(f => f.display_name === n || f.name.toUpperCase() === n.toUpperCase())).filter(Boolean),
    ...folders.filter(f => !folderOrder.some(n => f.display_name === n || f.name.toUpperCase() === n.toUpperCase())),
  ] as MailFolder[]

  return (
    <div className="flex h-screen bg-gray-100 overflow-hidden text-sm">
      {/* Sidebar */}
      <aside className="w-56 bg-white border-r border-gray-200 flex flex-col flex-shrink-0">
        <div className="px-4 py-3 border-b border-gray-100">
          <div className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">Mailbox</div>
          <select
            value={mailbox}
            onChange={e => switchMailbox(e.target.value)}
            className="w-full text-sm font-medium text-gray-800 bg-transparent focus:outline-none truncate"
          >
            {mailboxes.map(m => <option key={m} value={m}>{m}</option>)}
          </select>
        </div>

        <div className="px-3 py-3">
          <button
            onClick={() => { setReplyMsg(undefined); setCompose(true) }}
            className="w-full h-9 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-semibold text-sm transition-colors flex items-center justify-center gap-1.5"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Compose
          </button>
        </div>

        <nav className="flex-1 px-2 pb-3 overflow-y-auto">
          {sortedFolders.map(f => (
            <button
              key={f.name}
              onClick={() => { setActiveFolder(f.name); setPage(1) }}
              className={`w-full flex items-center justify-between px-3 py-2 rounded-xl text-sm transition-colors ${activeFolder === f.name ? 'bg-blue-50 text-blue-700 font-medium' : 'text-gray-600 hover:bg-gray-100'}`}
            >
              <span>{f.display_name}</span>
              {f.unread_count > 0 && (
                <span className="text-xs bg-blue-600 text-white rounded-full px-1.5 py-0.5 font-semibold">{f.unread_count}</span>
              )}
            </button>
          ))}
        </nav>
      </aside>

      {/* Message list */}
      <div className="w-72 bg-white border-r border-gray-200 flex flex-col flex-shrink-0">
        <div className="px-4 py-3 border-b border-gray-100 flex items-center justify-between">
          <span className="font-semibold text-gray-800">
            {sortedFolders.find(f => f.name === activeFolder)?.display_name || activeFolder}
          </span>
          <span className="text-xs text-gray-400">{total}</span>
        </div>
        <div className="flex-1 overflow-y-auto divide-y divide-gray-50">
          {loadingMsgs ? (
            <div className="flex items-center justify-center h-32">
              <div className="animate-spin w-5 h-5 rounded-full border-2 border-blue-600 border-t-transparent" />
            </div>
          ) : messages.length === 0 ? (
            <div className="text-center text-gray-400 py-12 text-sm">No messages</div>
          ) : messages.map(msg => (
            <button
              key={msg.uid}
              onClick={() => openMessage(msg)}
              className={`w-full text-left px-4 py-3 transition-colors ${selected?.uid === msg.uid ? 'bg-blue-50' : 'hover:bg-gray-50'}`}
            >
              <div className="flex items-center justify-between mb-0.5">
                <span className={`truncate max-w-[140px] ${!msg.is_read ? 'font-semibold text-gray-900' : 'text-gray-600'}`}>
                  {senderName(msg.from)}
                </span>
                <span className="text-xs text-gray-400 flex-shrink-0 ml-1">{formatDate(msg.date)}</span>
              </div>
              <p className={`truncate text-xs ${!msg.is_read ? 'text-gray-700 font-medium' : 'text-gray-500'}`}>{msg.subject || '(no subject)'}</p>
            </button>
          ))}
        </div>
        {total > 50 && (
          <div className="px-4 py-2 border-t border-gray-100 flex items-center justify-between">
            <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1}
              className="text-xs text-blue-600 disabled:text-gray-300">← Prev</button>
            <span className="text-xs text-gray-400">Page {page}</span>
            <button onClick={() => setPage(p => p + 1)} disabled={page * 50 >= total}
              className="text-xs text-blue-600 disabled:text-gray-300">Next →</button>
          </div>
        )}
      </div>

      {/* Message reader */}
      <main className="flex-1 bg-white overflow-y-auto">
        {!full && !selected && (
          <div className="flex flex-col items-center justify-center h-full text-gray-400">
            <svg className="w-12 h-12 mb-3 opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
            <p className="text-sm">Select a message to read</p>
          </div>
        )}
        {selected && !full && (
          <div className="flex items-center justify-center h-full">
            <div className="animate-spin w-6 h-6 rounded-full border-2 border-blue-600 border-t-transparent" />
          </div>
        )}
        {full && (
          <div className="max-w-3xl mx-auto p-8">
            <div className="flex items-start justify-between mb-6">
              <div className="flex-1 min-w-0">
                <h2 className="text-xl font-bold text-gray-900 mb-3">{full.subject || '(no subject)'}</h2>
                <div className="text-sm text-gray-600 space-y-1">
                  <div><span className="text-gray-400 w-8 inline-block">From</span> {full.from}</div>
                  <div><span className="text-gray-400 w-8 inline-block">To</span> {full.to?.join(', ')}</div>
                  {full.cc?.length > 0 && <div><span className="text-gray-400 w-8 inline-block">CC</span> {full.cc.join(', ')}</div>}
                  <div><span className="text-gray-400 w-8 inline-block">Date</span> {new Date(full.date).toLocaleString()}</div>
                </div>
              </div>
              <div className="flex gap-2 ml-4 flex-shrink-0">
                <button
                  onClick={() => { setReplyMsg(full); setCompose(true) }}
                  className="h-8 px-3 rounded-lg border border-gray-200 text-xs font-medium text-gray-600 hover:bg-gray-50 flex items-center gap-1"
                >
                  <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                  </svg>
                  Reply
                </button>
                <button
                  onClick={() => deleteMessage(full)}
                  className="h-8 px-3 rounded-lg border border-red-200 text-xs font-medium text-red-600 hover:bg-red-50"
                >
                  Delete
                </button>
              </div>
            </div>
            <div className="border-t border-gray-100 pt-6">
              {full.body_html ? (
                <iframe
                  srcDoc={full.body_html}
                  sandbox="allow-same-origin"
                  className="w-full border-0 min-h-96"
                  style={{ height: '60vh' }}
                  title="email"
                />
              ) : (
                <pre className="whitespace-pre-wrap text-sm text-gray-700 font-sans leading-relaxed">{full.body_text}</pre>
              )}
            </div>
          </div>
        )}
      </main>

      {(compose || replyMsg) && (
        <ComposeModal
          from={mailbox}
          replyTo={replyMsg}
          onClose={() => { setCompose(false); setReplyMsg(undefined) }}
          onSend={sendMessage}
        />
      )}
    </div>
  )
}

function WebmailPageInner() {
  const params = useSearchParams()
  const [password, setPassword] = useState<string | null>(null)
  const [allMailboxes, setAllMailboxes] = useState<string[]>([])

  const mailbox = params.get('mailbox') || ''

  useEffect(() => {
    const stored = sessionStorage.getItem(`mailbox_pass_${mailbox}`)
    if (stored) setPassword(stored)

    try {
      const raw = sessionStorage.getItem('webmail_mailboxes')
      if (raw) setAllMailboxes(JSON.parse(raw))
      else if (mailbox) setAllMailboxes([mailbox])
    } catch {
      if (mailbox) setAllMailboxes([mailbox])
    }
  }, [mailbox])

  if (!mailbox) {
    return (
      <div className="min-h-screen flex items-center justify-center text-gray-500 text-sm">
        No mailbox specified. Open from the Mailboxes page.
      </div>
    )
  }

  if (!password) {
    return (
      <UnlockScreen
        mailbox={mailbox}
        onUnlock={(p) => {
          sessionStorage.setItem(`mailbox_pass_${mailbox}`, p)
          setPassword(p)
        }}
      />
    )
  }

  return (
    <WebmailApp
      mailboxes={allMailboxes.length > 0 ? allMailboxes : [mailbox]}
      initialMailbox={mailbox}
      password={password}
    />
  )
}

export default function WebmailPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center"><div className="animate-spin w-6 h-6 rounded-full border-2 border-blue-600 border-t-transparent" /></div>}>
      <WebmailPageInner />
    </Suspense>
  )
}
