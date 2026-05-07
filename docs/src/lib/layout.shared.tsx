import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';
import { gitConfig } from './shared';
import { Nav } from '@/app/(home)/page';

type BaseOptionsConfig = {
  customNav?: boolean;
};

export function baseOptions({ customNav = true }: BaseOptionsConfig = {}): BaseLayoutProps {
  return {
    nav: customNav
      ? { component: <Nav /> }
      : {
        title: (
          <span>
            <span className="font-mono font-bold text-(--green)">CookieFarm</span>
          </span>
        ),
        url: '/',
      },
    themeSwitch: {
      enabled: false
    },

    githubUrl: `https://github.com/${gitConfig.user}/${gitConfig.repo}`,
  };
}
