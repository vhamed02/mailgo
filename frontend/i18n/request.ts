import { getRequestConfig } from 'next-intl/server'
import { cookies } from 'next/headers'
import { defaultLocale, isLocale, LOCALE_COOKIE } from './config'

// No URL-based routing: the active locale is stored in a cookie and defaults
// to Armenian. Falls back to the default when the cookie is missing/invalid.
export default getRequestConfig(async () => {
  const cookieLocale = cookies().get(LOCALE_COOKIE)?.value
  const locale = isLocale(cookieLocale) ? cookieLocale : defaultLocale

  return {
    locale,
    messages: (await import(`../messages/${locale}.json`)).default,
  }
})
