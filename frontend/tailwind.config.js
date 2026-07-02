/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      keyframes: {
        'fade-in': { from: { opacity: '0' }, to: { opacity: '1' } },
        'fade-out': { from: { opacity: '1' }, to: { opacity: '0' } },
        'fade-up': {
          '0%': { opacity: '0', transform: 'translateY(14px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        'zoom-in-95': { from: { opacity: '0', transform: 'scale(0.95)' }, to: { opacity: '1', transform: 'scale(1)' } },
        'zoom-out-95': { from: { opacity: '1', transform: 'scale(1)' }, to: { opacity: '0', transform: 'scale(0.95)' } },
        'slide-in-from-top': { from: { transform: 'translate(-50%, -48%)' }, to: { transform: 'translate(-50%, -50%)' } },
        'slide-out-to-top': { from: { transform: 'translate(-50%, -50%)' }, to: { transform: 'translate(-50%, -48%)' } },
      },
      animation: {
        'in': 'fade-in 0.15s ease-out',
        'out': 'fade-out 0.15s ease-in',
        'fade-up': 'fade-up 0.6s ease-out both',
        'zoom-in': 'zoom-in-95 0.2s ease-out',
        'zoom-out': 'zoom-out-95 0.15s ease-in',
        'slide-in': 'slide-in-from-top 0.2s ease-out',
        'slide-out': 'slide-out-to-top 0.15s ease-in',
      },
      colors: {
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))',
        },
        destructive: {
          DEFAULT: 'hsl(var(--destructive))',
          foreground: 'hsl(var(--destructive-foreground))',
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))',
        },
        accent: {
          DEFAULT: 'hsl(var(--accent))',
          foreground: 'hsl(var(--accent-foreground))',
        },
      },
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
      },
    },
  },
  plugins: [],
}
