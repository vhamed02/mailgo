import { cn } from '@/lib/utils'

type SectionProps = {
  id?: string
  as?: React.ElementType
  className?: string
  children: React.ReactNode
}

export function Section({ id, as: Tag = 'section', className, children }: SectionProps) {
  return (
    <Tag
      id={id}
      className={cn('scroll-mt-24 py-20 sm:py-24 lg:py-32', className)}
    >
      {children}
    </Tag>
  )
}
