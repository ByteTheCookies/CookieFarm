'use client';
import type React from 'react';
import { Inter } from 'next/font/google';
import './globals.css';
import { SidebarProvider } from '@/components/ui/sidebar';
import { AppSidebar } from '@/components/app-sidebar';
import { usePathname } from 'next/navigation';

const inter = Inter({ subsets: ['latin'] });

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const isLoginPage = pathname === '/login';
  const isConfigPage = pathname === '/config';

  return (
    <html lang="en" suppressHydrationWarning>
      <body className={inter.className}>
        {isLoginPage || isConfigPage ? (
          <main className="min-h-screen w-full">{children}</main>
        ) : (
          <SidebarProvider>
            <div className="flex min-h-screen w-full">
              <AppSidebar />
              <main className="flex-1 overflow-auto">{children}</main>
            </div>
          </SidebarProvider>
        )}
      </body>
    </html>
  );
}
