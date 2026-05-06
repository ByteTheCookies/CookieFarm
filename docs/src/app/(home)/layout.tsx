import { HomeLayout } from 'fumadocs-ui/layouts/home';
import { baseOptions } from '@/lib/layout.shared';
import { BookIcon, DownloadIcon, SyringeIcon } from 'lucide-react';
import { LinkItemType } from 'fumadocs-ui/layouts/shared';

export default function Layout({ children }: LayoutProps<'/'>) {
  return <HomeLayout
    links={[
      {
        text: 'Documentation',
        label: 'main docs',
        url: '/docs',
        icon: <BookIcon />,
        active: 'nested-url'
      },
      {
        text: 'Installation',
        label: 'installation docs',
        url: `/docs/installation`,
        icon: <DownloadIcon />,
        active: 'nested-url'
      },
      {
        text: 'Exploits',
        label: 'exploit docs',
        url: `/docs/exploits`,
        icon: <SyringeIcon />,
        active: 'nested-url'
      },
    ] as LinkItemType[]
    } {...baseOptions()}> {children}</HomeLayout >;
}
