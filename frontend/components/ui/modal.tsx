'use client'

import * as React from 'react'
import * as Dialog from '@radix-ui/react-dialog'
import { cn } from '@/lib/utils'

/* ── Root re-exports ─────────────────────────────────────────────────── */
export const Modal = Dialog.Root
export const ModalTrigger = Dialog.Trigger
export const ModalClose = Dialog.Close

/* ── Overlay ─────────────────────────────────────────────────────────── */
export const ModalOverlay = React.forwardRef<
  React.ElementRef<typeof Dialog.Overlay>,
  React.ComponentPropsWithoutRef<typeof Dialog.Overlay>
>(({ className, ...props }, ref) => (
  <Dialog.Overlay
    ref={ref}
    className={cn(
      'fixed inset-0 z-50 bg-black/50 backdrop-blur-sm',
      'animate-in', // custom keyframe defined in tailwind.config.js
      className
    )}
    {...props}
  />
))
ModalOverlay.displayName = 'ModalOverlay'

/* ── Content ─────────────────────────────────────────────────────────── */
export const ModalContent = React.forwardRef<
  React.ElementRef<typeof Dialog.Content>,
  React.ComponentPropsWithoutRef<typeof Dialog.Content>
>(({ className, children, ...props }, ref) => (
  <Dialog.Portal>
    <ModalOverlay />
    <Dialog.Content
      ref={ref}
      className={cn(
        'fixed left-1/2 top-1/2 z-50 -translate-x-1/2 -translate-y-1/2',
        'w-full max-w-lg rounded-2xl bg-white p-6 shadow-xl',
        'animate-zoom-in animate-slide-in',
        className
      )}
      {...props}
    >
      {children}

      {/* Close button */}
      <Dialog.Close
        className={cn(
          'absolute right-4 top-4 rounded-lg p-1.5 text-gray-400',
          'hover:bg-gray-100 hover:text-gray-600',
          'focus:outline-none focus:ring-2 focus:ring-gray-300',
          'transition-colors'
        )}
        aria-label="Close"
      >
        <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
        </svg>
      </Dialog.Close>
    </Dialog.Content>
  </Dialog.Portal>
))
ModalContent.displayName = 'ModalContent'

/* ── Typography helpers ──────────────────────────────────────────────── */
export const ModalHeader = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('mb-4 pr-6', className)} {...props} />
)

export const ModalTitle = React.forwardRef<
  React.ElementRef<typeof Dialog.Title>,
  React.ComponentPropsWithoutRef<typeof Dialog.Title>
>(({ className, ...props }, ref) => (
  <Dialog.Title
    ref={ref}
    className={cn('text-xl font-semibold text-gray-900', className)}
    {...props}
  />
))
ModalTitle.displayName = 'ModalTitle'

export const ModalDescription = React.forwardRef<
  React.ElementRef<typeof Dialog.Description>,
  React.ComponentPropsWithoutRef<typeof Dialog.Description>
>(({ className, ...props }, ref) => (
  <Dialog.Description
    ref={ref}
    className={cn('text-sm text-gray-500 mt-1', className)}
    {...props}
  />
))
ModalDescription.displayName = 'ModalDescription'

export const ModalFooter = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('mt-6 flex justify-end gap-3', className)} {...props} />
)
