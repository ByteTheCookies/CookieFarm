import type { NextConfig } from "next";

const isProduction = process.env.NODE_ENV === 'production';

const nextConfig: NextConfig = isProduction ? {
  output: 'standalone',
  experimental: {
    optimizeCss: true,
  },
  compress: true,
  images: {
    unoptimized: false,
  },
} : {
  experimental: {
    optimizeCss: false,
  },
  compress: false,
  images: {
    unoptimized: true,
  },
};

export default nextConfig;
