import {
  Navbar,
  Hero,
  LogoCloud,
  Features,
  ValueProposition,
  Stats,
  Testimonials,
  CTA,
  Footer,
} from '@/components/landing'

export default function Home() {
  return (
    <>
      <Navbar />
      <main>
        <Hero />
        <LogoCloud />
        <Features />
        <ValueProposition />
        <Stats />
        <Testimonials />
        <CTA />
      </main>
      <Footer />
    </>
  )
}
