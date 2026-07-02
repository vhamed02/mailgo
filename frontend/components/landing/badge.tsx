import { cn } from '@/lib/utils'

type BadgeProps = {
  className?: string
  children: React.ReactNode
}

export function Badge({ className, children }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-2 rounded-full border border-indigo-100 bg-indigo-50/80 px-3 py-1 text-xs font-semibold uppercase tracking-wider text-indigo-700',
        className
      )}
    >
      {children}
    </span>
  )
}
