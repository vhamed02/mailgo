import { cn } from '@/lib/utils'
import { Badge } from './badge'

type SectionHeadingProps = {
  eyebrow?: string
  title: string
  description?: string
  align?: 'center' | 'left'
  className?: string
}

export function SectionHeading({
  eyebrow,
  title,
  description,
  align = 'center',
  className,
}: SectionHeadingProps) {
  return (
    <div
      className={cn(
        'flex flex-col gap-4',
        align === 'center'
          ? 'mx-auto max-w-3xl items-center text-center'
          : 'max-w-2xl items-start text-left',
        className
      )}
    >
      {eyebrow ? <Badge>{eyebrow}</Badge> : null}
      <h2 className="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl lg:text-5xl">
        {title}
      </h2>
      {description ? (
        <p className="text-lg leading-relaxed text-slate-600 sm:text-xl">
          {description}
        </p>
      ) : null}
    </div>
  )
}
