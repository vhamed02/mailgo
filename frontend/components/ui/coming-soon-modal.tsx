'use client'

import {
  Modal,
  ModalContent,
  ModalHeader,
  ModalTitle,
  ModalDescription,
  ModalFooter,
  ModalClose,
} from '@/components/ui/modal'
import { Button } from '@/components/ui/button'

export interface ComingSoonFeature {
  title: string
  description: string
  /** Accent colour class applied to the icon container, e.g. "bg-blue-100" */
  iconBg?: string
  /** Colour class for the icon itself, e.g. "text-blue-600" */
  iconColor?: string
  /** SVG path(s) rendered inside a 24×24 viewBox */
  iconPath: React.ReactNode
  /** Bullet-point list of upcoming capabilities */
  items: string[]
}

interface ComingSoonModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  feature: ComingSoonFeature
}

export function ComingSoonModal({ open, onOpenChange, feature }: ComingSoonModalProps) {
  return (
    <Modal open={open} onOpenChange={onOpenChange}>
      <ModalContent>
        <ModalHeader>
          {/* Icon + title row */}
          <div className="flex items-center gap-4 mb-3">
            <div
              className={`w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0 ${feature.iconBg ?? 'bg-gray-100'}`}
            >
              <svg
                className={`w-6 h-6 ${feature.iconColor ?? 'text-gray-600'}`}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                {feature.iconPath}
              </svg>
            </div>

            <div>
              {/* 🚧 badge */}
              <span className="inline-flex items-center gap-1 text-xs font-medium text-amber-700 bg-amber-100 px-2 py-0.5 rounded-full mb-1">
                <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"
                  />
                </svg>
                Coming soon
              </span>

              <ModalTitle>{feature.title}</ModalTitle>
            </div>
          </div>

          <ModalDescription>{feature.description}</ModalDescription>
        </ModalHeader>

        {/* Feature list */}
        <ul className="space-y-2.5 mb-2">
          {feature.items.map((item) => (
            <li key={item} className="flex items-start gap-3 text-sm text-gray-700">
              <span
                className={`mt-0.5 w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0 ${feature.iconBg ?? 'bg-gray-100'}`}
              >
                <svg
                  className={`w-3 h-3 ${feature.iconColor ?? 'text-gray-600'}`}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M5 13l4 4L19 7" />
                </svg>
              </span>
              {item}
            </li>
          ))}
        </ul>

        <ModalFooter>
          <ModalClose asChild>
            <Button variant="outline" className="rounded-lg">
              Got it
            </Button>
          </ModalClose>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}
