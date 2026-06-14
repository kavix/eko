import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: "http://localhost:7700/api/:path*",
      },
    ];
  },
};

export default nextConfig;
