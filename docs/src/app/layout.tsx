import { RootProvider } from 'fumadocs-ui/provider/next';
import './global.css';
import type { Metadata } from 'next';
import { Geist, Geist_Mono, Inter } from 'next/font/google'

const _geist = Geist({ subsets: ["latin"] });
const _geistMono = Geist_Mono({ subsets: ["latin"] });

const inter = Inter({
  subsets: ['latin'],
});

export const metadata: Metadata = {
  title: {
    default: 'CookieFarm Documentation',
    template: '%s | CookieFarm Documentation',
  },
  description: 'Comprehensive documentation for CookieFarm, the ultimate tool for cookie management and exploitation.',
  keywords: ['CookieFarm', 'Documentation', 'CTF Framework', 'Cookie Management', 'Exploitation'],
  authors: [{ name: 'ByTheCookies' }],
  creator: 'ByTheCookies',
  publisher: 'ByTheCookies',
  metadataBase: new URL('https://cookiefarm.bytethecookies.org'),
  alternates: {
    canonical: '/',
  },
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: 'https://cookiefarm.bytethecookies.org',
    title: 'CookieFarm Documentation',
    description: 'Comprehensive documentation for CookieFarm, the ultimate tool for cookie management and exploitation.',
    siteName: 'My Website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'CookieFarm Documentation',
    description: 'Comprehensive documentation for CookieFarm, the ultimate tool for cookie management and exploitation.',
    creator: '@bytethecookies',
  },
  robots: {
    index: true,
    follow: true,
  },
  icons: {
    icon: '/favicon.ico',
    shortcut: '/favicon-16x16.png',
    apple: '/apple-touch-icon.png',
  },
};

export default function Layout({ children }: LayoutProps<'/'>) {
  return (
    <html lang="en" className={inter.className} suppressHydrationWarning>
      <body className="flex flex-col min-h-screen">
        <RootProvider theme={{
          defaultTheme: 'system',
        }}>
          {children}
        </RootProvider>
      </body>
    </html>
  );
}
