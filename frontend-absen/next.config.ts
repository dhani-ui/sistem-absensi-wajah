import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  webpack: (config, { isServer }) => {
    if (!isServer) {
      config.resolve.fallback = {
        ...config.resolve.fallback,
        fs: false,
      };
    }
    return config;
  },
  // eslint dihapus karena sudah tidak didukung di sini
  typescript: {
    ignoreBuildErrors: true, // Ini masih didukung untuk melewati error TypeScript
  },
};

export default nextConfig;
