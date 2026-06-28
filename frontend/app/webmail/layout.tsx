export default function WebmailLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-gray-100 h-screen overflow-hidden">{children}</body>
    </html>
  )
}
