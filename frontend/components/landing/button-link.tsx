import Link from 'next/link'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const landingButtonVariants = cva(
  'inline-flex items-center justify-center gap-2 rounded-xl font-semibold ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        primary:
          'bg-indigo-600 text-white hover:bg-indigo-700 shadow-lg shadow-indigo-600/20 hover:shadow-indigo-600/30',
        secondary:
          'bg-white text-slate-900 border border-slate-200 hover:bg-slate-50 hover:border-slate-300 shadow-sm',
        ghost: 'text-slate-700 hover:bg-slate-100',
        light:
          'bg-white text-indigo-700 hover:bg-indigo-50 shadow-lg shadow-indigo-900/10',
      },
      size: {
        md: 'h-11 px-5 text-sm',
        lg: 'h-12 px-7 text-base',
        xl: 'h-14 px-8 text-base',
      },
    },
    defaultVariants: {
      variant: 'primary',
      size: 'lg',
    },
  }
)

type ButtonLinkProps = {
  href: string
  className?: string
  children: React.ReactNode
} & Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'href'> &
  VariantProps<typeof landingButtonVariants>

export function ButtonLink({
  href,
  variant,
  size,
  className,
  children,
  ...props
}: ButtonLinkProps) {
  return (
    <Link
      href={href}
      className={cn(landingButtonVariants({ variant, size }), className)}
      {...props}
    >
      {children}
    </Link>
  )
}

export { landingButtonVariants }
