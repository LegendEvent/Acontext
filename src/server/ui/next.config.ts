import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";

const withNextIntl = createNextIntlPlugin();

const nextConfig: NextConfig = {
  /* config options here */
  output: "standalone",
  basePath: process.env.NEXT_PUBLIC_BASE_PATH || "",
};

export default withNextIntl(nextConfig);
