import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { Providers } from './providers'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: {
    default: 'MailGo — Professional email hosting for modern teams',
    template: '%s · MailGo',
  },
  description:
    'MailGo is professional email hosting with a lightning-fast webmail, admin controls and built-in deliverability. Bring your domain and get started in minutes.',
  keywords: [
    'email hosting',
    'business email',
    'custom domain email',
    'webmail',
    'professional email',
  ],
  openGraph: {
    title: 'MailGo — Professional email hosting for modern teams',
    description:
      'Secure, scalable email hosting with a fast webmail, admin dashboard and built-in deliverability tools.',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'MailGo — Professional email hosting for modern teams',
    description:
      'Secure, scalable email hosting with a fast webmail, admin dashboard and built-in deliverability tools.',
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <Providers>
          {children}
        </Providers>
      </body>
    </html>
  )
}
