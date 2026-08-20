import type { NextConfig } from "next";
import { dirname } from "node:path";
import { fileURLToPath } from "node:url";

const projectRoot = dirname(fileURLToPath(import.meta.url));

const nextConfig: NextConfig = {
  output: "standalone",
  turbopack: {
    root: projectRoot,
  },
  logging: {
    fetches: {
      fullUrl: true,
    },
    incomingRequests: {
      ignore: [/\/api\/health/],
    },
  },
};

export default nextConfig;
