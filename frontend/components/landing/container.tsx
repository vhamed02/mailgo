import { cn } from '@/lib/utils'

type ContainerProps = {
  as?: React.ElementType
  className?: string
  children: React.ReactNode
}

export function Container({ as: Tag = 'div', className, children }: ContainerProps) {
  return (
    <Tag className={cn('mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8', className)}>
      {children}
    </Tag>
  )
}
