import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

const BYTE_UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']

// Scales to the largest unit that keeps the value >= 1, with one decimal place
// unless it lands on a whole number (2 GB, not 2.0 GB). Decimal separator
// follows the viewer's locale, so hy/ru render "1,5 KB".
export function formatBytes(bytes: number) {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B'

  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < BYTE_UNITS.length - 1) {
    value /= 1024
    unit++
  }
  // Rounding can push a value back up into the next unit (1023.99 KB -> 1 MB),
  // which lands exactly on that unit's boundary.
  if (unit < BYTE_UNITS.length - 1 && Math.round(value * 10) / 10 >= 1024) {
    value = 1
    unit++
  }

  const decimals = unit === 0 || Number.isInteger(value) ? 0 : 1
  const text = value.toLocaleString(undefined, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
    useGrouping: false, // "1023 B", never "1,023 B"
  })
  return `${text} ${BYTE_UNITS[unit]}`
}
