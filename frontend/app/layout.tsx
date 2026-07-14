import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import { NextIntlClientProvider } from 'next-intl'
import { getLocale, getMessages } from 'next-intl/server'
import './globals.css'
import { Providers } from './providers'

const inter = Inter({ subsets: ['latin', 'cyrillic'] })

export const metadata: Metadata = {
  title: {
    default: 'Mailbox — Professional email hosting for modern teams',
    template: '%s · Mailbox',
  },
  description:
    'Mailbox is professional email hosting with a lightning-fast webmail, admin controls and built-in deliverability. Bring your domain and get started in minutes.',
  keywords: [
    'email hosting',
    'business email',
    'custom domain email',
    'webmail',
    'professional email',
  ],
  openGraph: {
    title: 'Mailbox — Professional email hosting for modern teams',
    description:
      'Secure, scalable email hosting with a fast webmail, admin dashboard and built-in deliverability tools.',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Mailbox — Professional email hosting for modern teams',
    description:
      'Secure, scalable email hosting with a fast webmail, admin dashboard and built-in deliverability tools.',
  },
}

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const locale = await getLocale()
  const messages = await getMessages()

  return (
    <html lang={locale}>
      <body className={inter.className}>
        <NextIntlClientProvider locale={locale} messages={messages}>
          <Providers>
            {children}
          </Providers>
        </NextIntlClientProvider>
      </body>
    </html>
  )
}
