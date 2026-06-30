import axios from 'axios'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

export function createMailClient(address: string, password: string) {
  const client = axios.create({
    baseURL: API_URL,
    headers: {
      'Content-Type': 'application/json',
      'X-Mailbox-Address': address,
      'X-Mailbox-Password': password,
    },
  })

  client.interceptors.request.use((config) => {
    const token = localStorage.getItem('access_token')
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  })

  return {
    listFolders: async () => {
      const r = await client.get('/mail/folders')
      return r.data.data as MailFolder[]
    },

    listMessages: async (folder: string, page = 1, limit = 50) => {
      const r = await client.get(`/mail/folders/${encodeURIComponent(folder)}/messages`, {
        params: { page, limit },
      })
      return r.data as { data: MailMessage[]; total: number }
    },

    getMessage: async (uid: number, folder: string) => {
      const r = await client.get(`/mail/messages/${uid}`, { params: { folder } })
      return r.data as MailMessage
    },

    markRead: async (uid: number, folder: string, read: boolean) => {
      await client.patch(`/mail/messages/${uid}/read`, { read }, { params: { folder } })
    },

    deleteMessage: async (uid: number, folder: string) => {
      await client.delete(`/mail/messages/${uid}`, { params: { folder } })
    },

    compose: async (payload: ComposePayload) => {
      await client.post('/mail/compose', payload)
    },

    reply: async (payload: ComposePayload) => {
      await client.post('/mail/reply', payload)
    },
  }
}

export type MailFolder = {
  name: string
  display_name: string
  unread_count: number
  total_count: number
}

export type MailMessage = {
  uid: number
  message_id: string
  in_reply_to?: string
  references?: string
  from: string
  to: string[]
  cc: string[]
  subject: string
  date: string
  is_read: boolean
  has_attachment: boolean
  folder?: string
  body_html?: string
  body_text?: string
  snippet?: string
}

export type ComposePayload = {
  to: string[]
  cc?: string[]
  subject: string
  body: string
  is_html?: boolean
  in_reply_to?: string
  references?: string
}
