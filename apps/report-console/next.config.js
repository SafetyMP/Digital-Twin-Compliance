const path = require("path");

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  transpilePackages: ["@digital-twin/console-shell"],
  outputFileTracingRoot: path.join(__dirname, "../.."),
};

module.exports = nextConfig;
