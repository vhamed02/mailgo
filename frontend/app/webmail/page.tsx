'use client'

import { useEffect, useState, useCallback, useRef, useMemo, Suspense } from 'react'
import { useSearchParams } from 'next/navigation'
import { createMailClient, type MailFolder, type MailMessage, type ComposePayload } from '@/lib/mail-api'

function formatDate(d: string) {
  const date = new Date(d)
  const now = new Date()
  if (date.toDateString() === now.toDateString())
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
}

function senderName(from: string) {
  const m = from.match(/^(.+?)\s*</)
  return m ? m[1].trim().replace(/^"(.*)"$/, '$1') : from.split('@')[0]
}

// Strip Re:/Fwd: prefixes (repeated, any locale-ish) to group a conversation.
function normalizeSubject(s: string) {
  let r = (s || '').trim()
  const re = /^(re|fwd|fw|aw|wg|sv|tr|rv):\s*/i
  while (re.test(r)) r = r.replace(re, '').trim()
  return r.toLowerCase()
}

type Conversation = {
  key: string
  subject: string
  messages: MailMessage[]      // newest-first (as returned by backend)
  unread: number
  latest: MailMessage
}

function UnlockScreen({ mailbox, onUnlock }: { mailbox: string; onUnlock: (p: string) => void }) {
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault(); setLoading(true); setError('')
    try {
      const client = createMailClient(mailbox, password)
      await client.listFolders()
      onUnlock(password)
    } catch { setError('Invalid password or mailbox unreachable.') }
    finally { setLoading(false) }
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
          <input type="password" placeholder="Mailbox password" value={password} onChange={e => setPassword(e.target.value)} required
            className="w-full h-11 px-4 rounded-xl border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
          {error && <p className="text-sm text-red-600">{error}</p>}
          <button type="submit" disabled={loading}
            className="w-full h-11 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-semibold text-sm transition-colors disabled:opacity-50">
            {loading ? 'Connecting...' : 'Open Webmail'}
          </button>
        </form>
      </div>
    </div>
  )
}

function ComposePanel({ from, replyTo, onSend, onDiscard }: {
  from: string
  replyTo?: MailMessage
  onSend: (p: ComposePayload) => Promise<void>
  onDiscard: () => void
}) {
  const [to, setTo] = useState(replyTo?.from ?? '')
  const [subject, setSubject] = useState(replyTo ? (replyTo.subject.startsWith('Re:') ? replyTo.subject : `Re: ${replyTo.subject}`) : '')
  const [body, setBody] = useState('')
  const [sending, setSending] = useState(false)
  const [error, setError] = useState('')
  const [sent, setSent] = useState(false)

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault(); setSending(true); setError('')
    try {
      await onSend({ to: to.split(',').map(s => s.trim()).filter(Boolean), subject, body, is_html: false, in_reply_to: replyTo?.message_id })
      setSent(true)
      setTimeout(onDiscard, 1200)
    } catch { setError('Failed to send. Please try again.') }
    finally { setSending(false) }
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-8 py-4 border-b border-gray-100">
        <h2 className="text-base font-semibold text-gray-900">{replyTo ? 'Reply' : 'New Message'}</h2>
        <button onClick={onDiscard} className="text-sm text-gray-400 hover:text-gray-600 transition-colors">Discard</button>
      </div>
      {sent ? (
        <div className="flex-1 flex flex-col items-center justify-center gap-3 text-green-600">
          <svg className="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p className="font-semibold text-lg">Message sent!</p>
        </div>
      ) : (
        <form onSubmit={handleSend} className="flex flex-col flex-1 overflow-hidden">
          <div className="px-8 py-0 divide-y divide-gray-100">
            <div className="flex items-center py-3 gap-8">
              <span className="text-xs font-semibold text-gray-400 uppercase w-12 flex-shrink-0">From</span>
              <span className="text-sm text-gray-600">{from}</span>
            </div>
            <div className="flex items-center py-3 gap-8">
              <span className="text-xs font-semibold text-gray-400 uppercase w-12 flex-shrink-0">To</span>
              <input value={to} onChange={e => setTo(e.target.value)} required placeholder="recipient@example.com"
                className="flex-1 text-sm text-gray-900 focus:outline-none bg-transparent" />
            </div>
            <div className="flex items-center py-3 gap-8">
              <span className="text-xs font-semibold text-gray-400 uppercase w-12 flex-shrink-0">Subject</span>
              <input value={subject} onChange={e => setSubject(e.target.value)} required placeholder="Subject"
                className="flex-1 text-sm text-gray-900 focus:outline-none bg-transparent" />
            </div>
          </div>
          <div className="flex-1 px-8 py-4 overflow-y-auto">
            {replyTo && (
              <div className="mb-4 pl-4 border-l-2 border-gray-200 text-xs text-gray-400 space-y-1">
                <p className="font-medium">On {new Date(replyTo.date).toLocaleString()}, {replyTo.from} wrote:</p>
                <p className="line-clamp-3">{replyTo.body_text || replyTo.snippet}</p>
              </div>
            )}
            <textarea
              value={body}
              onChange={e => setBody(e.target.value)}
              required
              placeholder="Write your message here..."
              className="w-full h-full min-h-64 text-sm text-gray-900 focus:outline-none resize-none bg-transparent leading-relaxed"
            />
          </div>
          {error && <p className="px-8 pb-2 text-xs text-red-600">{error}</p>}
          <div className="px-8 py-4 border-t border-gray-100 flex items-center gap-3">
            <button type="submit" disabled={sending}
              className="h-10 px-6 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-semibold text-sm transition-colors disabled:opacity-50 flex items-center gap-2">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
              </svg>
              {sending ? 'Sending...' : 'Send'}
            </button>
            <button type="button" onClick={onDiscard} className="h-10 px-4 rounded-xl text-sm text-gray-500 hover:bg-gray-100 transition-colors">Discard</button>
          </div>
        </form>
      )}
    </div>
  )
}

function MessageView({ msg, onReply, onDelete, compact }: {
  msg: MailMessage
  onReply: () => void
  onDelete: () => void
  compact?: boolean
}) {
  return (
    <div className={compact ? 'border-b border-gray-100 last:border-0 py-5' : 'py-6'}>
      <div className="flex items-start justify-between mb-4">
        <div className="flex-1 min-w-0">
          {!compact && <h2 className="text-xl font-bold text-gray-900 mb-3">{msg.subject || '(no subject)'}</h2>}
          <div className="text-sm text-gray-600 space-y-1">
            <div><span className="text-gray-400 w-12 inline-block">From</span>{msg.from}</div>
            <div><span className="text-gray-400 w-12 inline-block">To</span>{msg.to?.join(', ')}</div>
            {msg.cc?.length > 0 && <div><span className="text-gray-400 w-12 inline-block">CC</span>{msg.cc.join(', ')}</div>}
            <div><span className="text-gray-400 w-12 inline-block">Date</span>{new Date(msg.date).toLocaleString()}</div>
          </div>
        </div>
        <div className="flex gap-2 ml-4 flex-shrink-0">
          <button onClick={onReply}
            className="h-8 px-3 rounded-lg border border-gray-200 text-xs font-medium text-gray-600 hover:bg-gray-50 flex items-center gap-1.5 transition-colors">
            <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
            </svg>
            Reply
          </button>
          <button onClick={onDelete}
            className="h-8 px-3 rounded-lg border border-red-200 text-xs font-medium text-red-600 hover:bg-red-50 transition-colors">
            Delete
          </button>
        </div>
      </div>
      <div className="border-t border-gray-100 pt-4">
        {msg.body_html ? (
          <iframe srcDoc={msg.body_html} sandbox="allow-same-origin" className="w-full border-0" style={{ minHeight: compact ? '40vh' : '60vh' }} title="email" />
        ) : (
          <pre className="whitespace-pre-wrap text-sm text-gray-700 font-sans leading-relaxed">{msg.body_text}</pre>
        )}
      </div>
    </div>
  )
}

function WebmailApp({ mailboxes, initialMailbox, password }: {
  mailboxes: string[]; initialMailbox: string; password: string
}) {
  const [mailbox, setMailbox] = useState(initialMailbox)
  const [client, setClient] = useState(() => createMailClient(initialMailbox, password))
  const [folders, setFolders] = useState<MailFolder[]>([])
  const [activeFolder, setActiveFolder] = useState('INBOX')
  const [messages, setMessages] = useState<MailMessage[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [selectedKey, setSelectedKey] = useState<string | null>(null)
  const [thread, setThread] = useState<MailMessage[]>([])   // full messages, oldest→newest
  const [loadingMsgs, setLoadingMsgs] = useState(false)
  const [view, setView] = useState<'empty' | 'loading' | 'thread' | 'compose' | 'reply'>('empty')
  const [replyMsg, setReplyMsg] = useState<MailMessage | undefined>()

  const scrollRef = useRef<HTMLDivElement>(null)
  const pollingRef = useRef(false)

  const switchMailbox = (addr: string) => {
    setMailbox(addr); setClient(createMailClient(addr, password))
    setActiveFolder('INBOX'); setSelectedKey(null); setThread([]); setPage(1); setView('empty')
  }

  const loadFolders = useCallback(async () => {
    try { setFolders(await client.listFolders()) } catch {}
  }, [client])

  const loadMessages = useCallback(async (silent = false) => {
    if (!silent) setLoadingMsgs(true)
    try {
      const r = await client.listMessages(activeFolder, page)
      setMessages(r.data); setTotal(r.total)
    } catch {} finally { if (!silent) setLoadingMsgs(false) }
  }, [client, activeFolder, page])

  // Group messages into conversations by normalized subject.
  const conversations: Conversation[] = useMemo(() => {
    const map: Record<string, Conversation> = {}
    const ordered: Conversation[] = []
    for (const m of messages) {
      const key = normalizeSubject(m.subject) || '(no subject)'
      if (!map[key]) {
        const c: Conversation = { key, subject: m.subject || '(no subject)', messages: [], unread: 0, latest: m }
        map[key] = c; ordered.push(c)
      }
      map[key].messages.push(m)
      if (!m.is_read) map[key].unread++
      if (new Date(m.date) > new Date(map[key].latest.date)) map[key].latest = m
    }
    return ordered
  }, [messages])

  // Initial load (folder / page change) — reset selection.
  useEffect(() => { loadFolders() }, [loadFolders])
  useEffect(() => { loadMessages(); setSelectedKey(null); setThread([]); setView('empty') }, [loadMessages])

  // Auto-refresh inbox every 5 seconds (silent — does not disturb reading).
  useEffect(() => {
    const id = setInterval(async () => {
      if (pollingRef.current) return
      pollingRef.current = true
      try {
        // Refresh folder list (for unread badges) + message list silently.
        try { setFolders(await client.listFolders()) } catch {}
        await loadMessages(true)
        // If a conversation is open, refresh its full messages too so new
        // replies appear below automatically (like Gmail).
        if (selectedKey) {
          const c = conversations.find(c => c.key === selectedKey)
          if (c) await refreshThread(c, true)
        }
      } finally { pollingRef.current = false }
    }, 5000)
    return () => clearInterval(id)
  }, [client, loadMessages, selectedKey, conversations])

  // Refresh the full bodies of a conversation's messages (oldest→newest).
  const refreshThread = async (c: Conversation, silent = false) => {
    if (!silent) setView('loading')
    try {
      // c.messages is newest-first; reverse to oldest-first like Gmail thread.
      const ordered = [...c.messages].reverse()
      const fulls = await Promise.all(ordered.map(m => client.getMessage(m.uid, activeFolder)))
      fulls.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
      setThread(fulls)
      if (!silent) setView('thread')
      // Scroll the newest message into view (bottom) once rendered.
      setTimeout(() => {
        const el = scrollRef.current
        if (el) el.scrollTop = el.scrollHeight
      }, 50)
    } catch { if (!silent) setView('empty') }
  }

  const openConversation = async (c: Conversation) => {
    setSelectedKey(c.key)
    await refreshThread(c)
    // Mark every unread message in the conversation as read.
    const unreads = c.messages.filter(m => !m.is_read)
    for (const m of unreads) {
      try { await client.markRead(m.uid, activeFolder, true) } catch {}
    }
    // Reflect read state optimistically in the list + folder unread counts.
    if (unreads.length > 0) {
      setMessages(prev => prev.map(m =>
        normalizeSubject(m.subject) === c.key ? { ...m, is_read: true } : m))
      setFolders(prev => prev.map(f => f.name === activeFolder
        ? { ...f, unread_count: Math.max(0, f.unread_count - unreads.length) } : f))
    }
  }

  const deleteMessage = async (msg: MailMessage) => {
    try {
      await client.deleteMessage(msg.uid, activeFolder)
      setThread(prev => prev.filter(m => m.uid !== msg.uid))
      if (thread.filter(m => m.uid !== msg.uid).length === 0) {
        setSelectedKey(null); setView('empty')
      }
      loadMessages(true)
    } catch {}
  }

  const sendMessage = async (payload: ComposePayload) => {
    if (view === 'reply' && replyMsg) await client.reply(payload)
    else await client.compose(payload)
    loadMessages(true)
  }

  const folderOrder = ['INBOX', 'Sent', 'Drafts', 'Spam', 'Trash']
  const sortedFolders = [
    ...folderOrder.map(n => folders.find(f => f.display_name === n || f.name.toUpperCase() === n.toUpperCase())).filter(Boolean),
    ...folders.filter(f => !folderOrder.some(n => f.display_name === n || f.name.toUpperCase() === n.toUpperCase())),
  ] as MailFolder[]

  return (
    <div className="flex h-screen bg-gray-50 overflow-hidden text-sm">
      {/* Sidebar */}
      <aside className="w-56 bg-white border-r border-gray-200 flex flex-col flex-shrink-0">
        <div className="px-4 py-3 border-b border-gray-100">
          <div className="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-1">Mailbox</div>
          <select value={mailbox} onChange={e => switchMailbox(e.target.value)}
            className="w-full text-sm font-medium text-gray-800 bg-transparent focus:outline-none truncate">
            {mailboxes.map(m => <option key={m} value={m}>{m}</option>)}
          </select>
        </div>
        <div className="px-3 py-3">
          <button
            onClick={() => { setReplyMsg(undefined); setView('compose') }}
            className="w-full h-9 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-semibold text-sm transition-colors flex items-center justify-center gap-1.5">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Compose
          </button>
        </div>
        <nav className="flex-1 px-2 pb-3 overflow-y-auto">
          {sortedFolders.map(f => (
            <button key={f.name} onClick={() => { setActiveFolder(f.name); setPage(1) }}
              className={`w-full flex items-center justify-between px-3 py-2 rounded-xl text-sm transition-colors ${activeFolder === f.name ? 'bg-blue-50 text-blue-700 font-medium' : 'text-gray-600 hover:bg-gray-100'}`}>
              <span>{f.display_name}</span>
              {f.unread_count > 0 && <span className="text-xs bg-blue-600 text-white rounded-full px-1.5 py-0.5 font-semibold">{f.unread_count}</span>}
            </button>
          ))}
        </nav>
      </aside>

      {/* Message list */}
      <div className="w-72 bg-white border-r border-gray-200 flex flex-col flex-shrink-0">
        <div className="px-4 py-3 border-b border-gray-100 flex items-center justify-between">
          <span className="font-semibold text-gray-800">{sortedFolders.find(f => f.name === activeFolder)?.display_name || activeFolder}</span>
          <span className="text-xs text-gray-400">{total}</span>
        </div>
        <div className="flex-1 overflow-y-auto divide-y divide-gray-50">
          {loadingMsgs ? (
            <div className="flex items-center justify-center h-32">
              <div className="animate-spin w-5 h-5 rounded-full border-2 border-blue-600 border-t-transparent" />
            </div>
          ) : conversations.length === 0 ? (
            <div className="text-center text-gray-400 py-12 text-sm">No messages</div>
          ) : conversations.map(c => (
            <button key={c.key} onClick={() => openConversation(c)}
              className={`w-full text-left px-4 py-3 transition-colors ${selectedKey === c.key ? 'bg-blue-50' : 'hover:bg-gray-50'}`}>
              <div className="flex items-center justify-between mb-0.5">
                <span className={`truncate max-w-[140px] ${c.unread > 0 ? 'font-semibold text-gray-900' : 'text-gray-600'}`}>{senderName(c.latest.from)}</span>
                <span className="text-xs text-gray-400 flex-shrink-0 ml-1">{formatDate(c.latest.date)}</span>
              </div>
              <div className="flex items-center gap-1.5">
                <p className={`truncate text-xs flex-1 ${c.unread > 0 ? 'text-gray-700 font-medium' : 'text-gray-500'}`}>{c.subject || '(no subject)'}</p>
                {c.messages.length > 1 && <span className="text-[10px] px-1 py-0.5 rounded bg-gray-200 text-gray-500 font-medium flex-shrink-0">{c.messages.length}</span>}
                {c.unread > 0 && <span className="w-2 h-2 rounded-full bg-blue-600 flex-shrink-0" />}
              </div>
            </button>
          ))}
        </div>
        {total > 50 && (
          <div className="px-4 py-2 border-t border-gray-100 flex items-center justify-between">
            <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1} className="text-xs text-blue-600 disabled:text-gray-300">← Prev</button>
            <span className="text-xs text-gray-400">Page {page}</span>
            <button onClick={() => setPage(p => p + 1)} disabled={page * 50 >= total} className="text-xs text-blue-600 disabled:text-gray-300">Next →</button>
          </div>
        )}
      </div>

      {/* Main panel */}
      <main className="flex-1 bg-white overflow-y-auto flex flex-col">
        {view === 'empty' && (
          <div className="flex flex-col items-center justify-center h-full text-gray-400">
            <svg className="w-12 h-12 mb-3 opacity-20" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
            <p className="text-sm">Select a conversation to read</p>
          </div>
        )}

        {view === 'loading' && (
          <div className="flex items-center justify-center h-full">
            <div className="animate-spin w-6 h-6 rounded-full border-2 border-blue-600 border-t-transparent" />
          </div>
        )}

        {view === 'thread' && (
          <div className="max-w-3xl mx-auto w-full px-8 py-6 flex flex-col" style={{ minHeight: '100%' }}>
            <div className="flex items-center justify-between pb-4 border-b border-gray-100">
              <h2 className="text-xl font-bold text-gray-900">{thread[0]?.subject || '(no subject)'}</h2>
              <span className="text-xs text-gray-400">{thread.length} message{thread.length > 1 ? 's' : ''}</span>
            </div>
            <div ref={scrollRef} className="flex-1 overflow-y-auto">
              {thread.map((msg, i) => (
                <MessageView
                  key={msg.uid}
                  msg={msg}
                  compact={i > 0}
                  onReply={() => { setReplyMsg(msg); setView('reply') }}
                  onDelete={() => deleteMessage(msg)}
                />
              ))}
            </div>
          </div>
        )}

        {(view === 'compose' || view === 'reply') && (
          <ComposePanel
            from={mailbox}
            replyTo={view === 'reply' ? replyMsg : undefined}
            onSend={sendMessage}
            onDiscard={() => { setView(thread.length > 0 ? 'thread' : 'empty'); setReplyMsg(undefined) }}
          />
        )}
      </main>
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
      setAllMailboxes(raw ? JSON.parse(raw) : mailbox ? [mailbox] : [])
    } catch { if (mailbox) setAllMailboxes([mailbox]) }
  }, [mailbox])

  if (!mailbox) return <div className="min-h-screen flex items-center justify-center text-gray-500 text-sm">No mailbox specified.</div>
  if (!password) return <UnlockScreen mailbox={mailbox} onUnlock={p => { sessionStorage.setItem(`mailbox_pass_${mailbox}`, p); setPassword(p) }} />
  return <WebmailApp mailboxes={allMailboxes.length > 0 ? allMailboxes : [mailbox]} initialMailbox={mailbox} password={password} />
}

export default function WebmailPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center"><div className="animate-spin w-6 h-6 rounded-full border-2 border-blue-600 border-t-transparent" /></div>}>
      <WebmailPageInner />
    </Suspense>
  )
}
