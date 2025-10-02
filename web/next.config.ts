import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  images: {
    // Загрузка фото Лего для водителей
    domains: ["randomuser.me"],
  },
  reactStrictMode: false,
};

export default nextConfig;
