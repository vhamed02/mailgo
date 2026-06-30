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

// Gmail-style conversation threading.
//
// A Conversation (Thread) is an ORDERED LIST of independent, first-class
// messages — NEVER a parent/child tree. Threading is derived purely from RFC
// metadata (Message-ID, In-Reply-To, References). There is no parentId, no
// nested array, no recursive rendering.
//
// Two messages belong to the same thread if they share a common ancestor
// (transitively via References/In-Reply-To) or — pragmatic fallback for
// clients that don't set the headers — share a normalized subject.

function normalizeSubject(s: string) {
  let r = (s || '').trim()
  const re = /^(re|fwd|fw|aw|wg|sv|tr|rv):\s*/i
  while (re.test(r)) r = r.replace(re, '').trim()
  return r.toLowerCase()
}

// Split a References header (whitespace-separated <id@host>) into a clean list.
function refIds(refs: string | undefined): string[] {
  if (!refs) return []
  return refs.split(/\s+/).map(s => s.trim()).filter(Boolean)
}

type Conversation = {
  key: string               // stable thread root id
  subject: string
  messages: MailMessage[]   // newest-first (as returned by backend)
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
      await onSend({
        to: to.split(',').map(s => s.trim()).filter(Boolean),
        subject,
        body,
        is_html: false,
        in_reply_to: replyTo?.message_id,
        references: replyTo ? [replyTo.references, replyTo.message_id].filter(Boolean).join(' ') : undefined,
      })
      setSent(true)
      setTimeout(onDiscard, 800)
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

<<<<<<< Updated upstream
function MessageView({ msg, onReply, compact }: {
=======
// A segment of a reply chain extracted from a single email's body. Many email
// clients (Gmail, Apple Mail, Outlook) embed the quoted previous message as a
// <blockquote class="gmail_quote">. We split that block so each reply shows as
// its own row, exactly like Gmail's "message → reply → reply" stacked view.
type QuoteSegment = {
  from?: string
  date?: string
  bodyHtml: string
}

function parseQuoteChain(html: string, baseFrom?: string, baseDate?: string): QuoteSegment[] {
  if (!html) return baseFrom || baseDate ? [{ from: baseFrom, date: baseDate, bodyHtml: '' }] : []
  const out: QuoteSegment[] = []
  let currentHtml = html
  let curFrom = baseFrom
  let curDate = baseDate

  let depth = 0
  while (currentHtml && depth < 25) {
    depth++
    const doc = new DOMParser().parseFromString(currentHtml, 'text/html')
    let quoteContainer: Element | null = doc.querySelector('div.gmail_quote_container')
    let blockquote: Element | null = null
    if (quoteContainer) blockquote = quoteContainer.querySelector('blockquote.gmail_quote')
    if (!quoteContainer) {
      blockquote = doc.querySelector('blockquote.gmail_quote')
      if (blockquote) quoteContainer = blockquote
    }

    if (!quoteContainer) {
      out.push({ from: curFrom, date: curDate, bodyHtml: currentHtml.trim() })
      break
    }

    // Segment body = everything except the quote container.
    const cloned = doc.cloneNode(true) as Document
    const q = cloned.querySelector('div.gmail_quote_container') || cloned.querySelector('blockquote.gmail_quote')
    if (q && q.parentNode) q.parentNode.removeChild(q)
    let segBody = cloned.body ? cloned.body.innerHTML : ''
    segBody = segBody.replace(/(<br\s*\/?>\s*)+$/i, '').replace(/(<div\s*>\s*<\/div>\s*)+$/i, '').trim()
    out.push({ from: curFrom, date: curDate, bodyHtml: segBody })

    // Parse "On <date>, <sender> wrote:" for the older (next) message.
    const attr = quoteContainer.querySelector('.gmail_attr')
    let nextFrom: string | undefined
    let nextDate: string | undefined
    if (attr) {
      const txt = (attr.textContent || '').trim()
      const m = txt.match(/^\s*On\s+(.+?)\s+wrote:\s*$/is)
      if (m) {
        const inner = m[1].trim()
        const emailMatch = inner.match(/^(.*?)\s+(.+\s*<[^>]+>)\s*$/s)
        if (emailMatch) {
          nextDate = emailMatch[1].trim()
          nextFrom = emailMatch[2].trim()
        } else {
          nextDate = inner
        }
      }
    }

    if (blockquote) {
      currentHtml = blockquote.innerHTML
      curFrom = nextFrom || curFrom
      curDate = nextDate || curDate
    } else {
      if (nextFrom || nextDate) out.push({ from: nextFrom, date: nextDate, bodyHtml: '' })
      break
    }
  }
  return out
}

// Split a parsed "From" string into display name + email.
// "MailGo <support@yerevan.digital>" -> { name: "MailGo", email: "support@yerevan.digital" }
function splitFrom(s?: string): { name: string; email: string } {
  if (!s) return { name: '', email: '' }
  const m = s.match(/^\s*"?(.+?)"?\s*<([^>]+)>\s*$/)
  if (m) return { name: m[1].trim(), email: m[2].trim() }
  if (s.includes('@')) return { name: s.split('@')[0], email: s }
  return { name: s, email: '' }
}

// Extract the bare email address from a "Name <email>" / "email" string.
function stripEmail(s: string): string {
  const m = s.match(/<([^>]+)>/)
  if (m) return m[1]
  return s.replace(/^\s+|\s+$/g, '')
}

// One message rendered as a Gmail-style thread: a flat list of independent
// `role="listitem"` cards — oldest on top, newest at the bottom. No nesting,
// no tree, no recursive rendering. The oldest quoted replies embedded in the
// body (Gmail-style <blockquote class="gmail_quote">) are split out into their
// own first-class cards above; the newest message keeps its full body
// (including any embedded blockquote), exactly like Gmail.
function MessageCard({ msg, isLast, onReply, onDelete }: {
>>>>>>> Stashed changes
  msg: MailMessage
  isLast: boolean
  onReply: () => void
<<<<<<< Updated upstream
  compact?: boolean
=======
  onDelete: () => void
>>>>>>> Stashed changes
}) {
  const segments = useMemo(
    () => msg.body_html ? parseQuoteChain(msg.body_html, msg.from, msg.date) : [],
    [msg.body_html, msg.from, msg.date],
  )
  // segments[0] = newest (quote-stripped body); segments[1..] = progressively older.
  // For display order (oldest first), reverse the older slice, then append newest.
  const older = segments.slice(1).reverse()   // oldest first
  const newest = segments[0]
  const recipientLabel = msg.to?.length ? `to ${msg.to.map(senderName).join(', ')}` : ''
  const newestName = senderName(msg.from)
  const newestEmail = stripEmail(msg.from)

  return (
<<<<<<< Updated upstream
    <div className={compact ? 'border-t border-gray-100 pt-6 mt-6' : 'pt-2'}>
      {/* Header */}
      <div className="flex items-start justify-between gap-4 mb-4">
        <div className="flex-1 min-w-0">
          {!compact && (
            <h1 className="text-2xl font-bold text-gray-900 leading-tight mb-3">
              {msg.subject || '(no subject)'}
            </h1>
          )}
          <div className="flex flex-col gap-1 text-sm">
            <div className="flex items-baseline gap-2">
              <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide w-10 flex-shrink-0">From</span>
              <span className="text-gray-800 font-medium">{msg.from}</span>
            </div>
            {msg.to?.length > 0 && (
              <div className="flex items-baseline gap-2">
                <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide w-10 flex-shrink-0">To</span>
                <span className="text-gray-600">{msg.to.join(', ')}</span>
              </div>
            )}
            {msg.cc?.length > 0 && (
              <div className="flex items-baseline gap-2">
                <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide w-10 flex-shrink-0">CC</span>
                <span className="text-gray-600">{msg.cc.join(', ')}</span>
              </div>
            )}
            <div className="flex items-baseline gap-2">
              <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide w-10 flex-shrink-0">Date</span>
              <span className="text-gray-500 text-xs">{new Date(msg.date).toLocaleString([], { dateStyle: 'medium', timeStyle: 'short' })}</span>
=======
    <div role="list" className="divide-y divide-gray-100">
      {/* Older quoted messages — each a first-class card (oldest on top) */}
      {older.map((seg, i) => {
        const { name, email } = splitFrom(seg.from)
        const dateLabel = seg.date || ''
        return (
          <div role="listitem" key={`old-${i}`} className="py-5">
            <div className="message-header flex flex-wrap items-baseline gap-x-2 gap-y-0.5 mb-2">
              <span className="font-semibold text-gray-900">{name || 'Previous message'}</span>
              {email && <span className="text-xs text-gray-400">&lt;{email}&gt;</span>}
              {dateLabel && <span className="text-xs text-gray-400 ml-auto">{dateLabel}</span>}
            </div>
            <div className="message-body">
              {seg.bodyHtml ? (
                <iframe srcDoc={seg.bodyHtml} sandbox="allow-same-origin" className="w-full border-0" style={{ minHeight: '30vh' }} title={`old-${i}`} />
              ) : (
                <pre className="whitespace-pre-wrap text-sm text-gray-700 font-sans leading-relaxed">(no content)</pre>
              )}
>>>>>>> Stashed changes
            </div>
          </div>
        )
      })}

      {/* Newest message — body with embedded quote block stripped (already
          rendered as its own card above, exactly like Gmail). */}
      <div role="listitem" className={`py-5 ${isLast ? '' : ''}`}>
        <div className="message-header flex flex-wrap items-baseline gap-x-2 gap-y-0.5 mb-1">
          <span className="font-semibold text-gray-900">{newestName}</span>
          <span className="text-xs text-gray-400">&lt;{newestEmail}&gt;</span>
          <span className="text-xs text-gray-400 ml-auto">
            {new Date(msg.date).toLocaleString()}
            {recipientLabel && <span className="ml-2">· {recipientLabel}</span>}
          </span>
        </div>
<<<<<<< Updated upstream
        {/* Primary Reply CTA */}
        <button
          onClick={onReply}
          className="flex-shrink-0 h-9 px-4 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-semibold text-sm transition-colors flex items-center gap-1.5 shadow-sm"
        >
          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
          </svg>
          Reply
        </button>
      </div>

      {/* Body */}
      <div className={`${compact ? '' : 'border-t border-gray-100 pt-5'}`}>
        {msg.body_html ? (
          <iframe
            srcDoc={msg.body_html}
            sandbox="allow-same-origin"
            className="w-full border-0 rounded-lg"
            style={{ minHeight: compact ? '30vh' : '55vh' }}
            title="email"
          />
        ) : (
          <pre className="whitespace-pre-wrap text-sm text-gray-700 font-sans leading-relaxed">{msg.body_text}</pre>
        )}
=======
        <div className="flex gap-2 mb-3">
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
        <div className="message-body">
          {newest?.bodyHtml ? (
            <iframe srcDoc={newest.bodyHtml} sandbox="allow-same-origin" className="w-full border-0" style={{ minHeight: '50vh' }} title={`msg-${msg.uid}`} />
          ) : msg.body_html ? (
            <iframe srcDoc={msg.body_html} sandbox="allow-same-origin" className="w-full border-0" style={{ minHeight: '50vh' }} title={`msg-${msg.uid}`} />
          ) : (
            <pre className="whitespace-pre-wrap text-sm text-gray-700 font-sans leading-relaxed">{msg.body_text}</pre>
          )}
        </div>
>>>>>>> Stashed changes
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

  // Group messages into conversations by RFC threading (References/In-Reply-To),
  // with a normalized-subject fallback. The result is a flat ordered list of
  // independent messages per conversation — no parent/child tree.
  const conversations: Conversation[] = useMemo(() => {
    const byId: Record<string, MailMessage> = {}
    for (const m of messages) if (m.message_id) byId[m.message_id] = m

    // Union-Find over Message-IDs linking messages in the same thread.
    const parent: Record<string, string> = {}
    const find = (x: string): string => {
      if (parent[x] === undefined) parent[x] = x
      while (parent[x] !== x) { parent[x] = parent[parent[x]]; x = parent[x] }
      return x
    }
    const union = (a: string, b: string) => {
      const ra = find(a), rb = find(b)
      if (ra !== rb) parent[ra] = rb
    }

    for (const m of messages) {
      const a = m.message_id || `uid:${m.uid}`
      const ids = [...refIds(m.references)]
      if (m.in_reply_to) ids.push(m.in_reply_to)
      for (const r of ids) {
        // Only union with ancestors that actually exist in this mailbox window,
        // OR chain to the reference id itself (the root may be off-window).
        union(a, r)
      }
    }

    const buckets: Record<string, Conversation> = {}
    const ordered: Conversation[] = []
    for (const m of messages) {
      // Prefer RFC threading root; as a stable fallback for messages with no
      // References at all, fall back to normalized subject so user-visible
      // replies on the same topic still group.
      const root = (m.references || m.in_reply_to) ? find(m.message_id || `uid:${m.uid}`) : `subj:${normalizeSubject(m.subject)}`
      if (!buckets[root]) {
        const c: Conversation = { key: root, subject: m.subject || '(no subject)', messages: [], unread: 0, latest: m }
        buckets[root] = c; ordered.push(c)
      }
      buckets[root].messages.push(m)
      if (!m.is_read) buckets[root].unread++
      if (new Date(m.date) > new Date(buckets[root].latest.date)) buckets[root].latest = m
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
    if (view === 'reply' && replyMsg) {
      await client.reply(payload)
      // Optimistically append the sent reply to the current thread so it
      // appears immediately without waiting for the next poll cycle.
      const optimisticMsg: MailMessage = {
        uid: Date.now(), // temporary ID until next refresh
        message_id: '',
        from: mailbox,
        to: payload.to,
        cc: payload.cc ?? [],
        subject: payload.subject,
        date: new Date().toISOString(),
        is_read: true,
        has_attachment: false,
        folder: 'Sent',
        body_text: payload.body,
        body_html: payload.is_html ? payload.body : undefined,
        snippet: payload.body.slice(0, 100),
      }
      setThread(prev => [...prev, optimisticMsg])
      setTimeout(() => {
        const el = scrollRef.current
        if (el) el.scrollTop = el.scrollHeight
      }, 50)
    } else {
      await client.compose(payload)
    }
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
<<<<<<< Updated upstream
          <div className="max-w-3xl mx-auto w-full px-8 py-6 flex flex-col" style={{ minHeight: '100%' }}>
            <div className="pb-5 border-b border-gray-100">
              <h1 className="text-2xl font-bold text-gray-900 leading-tight">{thread[0]?.subject || '(no subject)'}</h1>
              <p className="text-xs text-gray-400 mt-1">{thread.length} message{thread.length > 1 ? 's' : ''}</p>
=======
          <>
          {thread[0]?.subject && (
            <div className="w-full px-8 pt-6 pb-3">
              <h2 className="text-xl font-bold text-gray-900">{thread[0].subject}</h2>
>>>>>>> Stashed changes
            </div>
          )}
          <div className="w-full px-8 py-6 flex flex-col" style={{ minHeight: '100%' }}>
            <div ref={scrollRef} className="flex-1 overflow-y-auto divide-y divide-gray-100">
              {/* Flat list of independent message cards — oldest first (Gmail). */}
              {thread.map((msg, i) => (
                <MessageCard
                  key={msg.uid}
                  msg={msg}
                  isLast={i === thread.length - 1}
                  onReply={() => { setReplyMsg(msg); setView('reply') }}
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
