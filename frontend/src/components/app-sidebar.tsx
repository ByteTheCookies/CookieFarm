'use client';

import { BarChart3, Shield, FileText, Settings } from 'lucide-react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar';
import Image from 'next/image';

const navigationItems = [
  {
    title: 'Match Overview',
    url: '/',
    icon: BarChart3,
  },
  {
    title: 'Service Monitoring',
    url: '/monitoring',
    icon: Shield,
  },
  {
    title: 'Flag Logs',
    url: '/logs',
    icon: FileText,
  },
];

const footerItems = [
  {
    title: 'Settings',
    url: '/settings',
    icon: Settings,
  },
];

export function AppSidebar() {
  const pathname = usePathname();

  return (
    <Sidebar className="border-border/40 border-r">
      <SidebarHeader className="border-border/40 border-b p-6">
        <div className="flex items-center justify-around gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-lg">
            <Image alt="logo" src={'images/logo.png'} width={50} height={20} />
          </div>
          <div className="text-center">
            <h1 className="text-foreground text-2xl font-bold">CookieFarm</h1>
            <p className="text-muted-foreground text-[0.7rem] leading-3.5">
              Zero distraction <br /> Only exploiting
            </p>
          </div>
        </div>
      </SidebarHeader>

      <SidebarContent className="px-4 py-6">
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {navigationItems.map(item => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    asChild
                    isActive={pathname === item.url}
                    className="w-full justify-start"
                  >
                    <Link href={item.url} className="flex items-center gap-3">
                      <item.icon className="h-5 w-5" />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="border-border/40 border-t p-4">
        <SidebarMenu>
          {footerItems.map(item => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton
                asChild
                isActive={pathname === item.url}
                className="w-full justify-start"
              >
                <Link href={item.url} className="flex items-center gap-3">
                  <item.icon className="h-5 w-5" />
                  <span>{item.title}</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  );
}
