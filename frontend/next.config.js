// /** @type {import('next').NextConfig} */
// const nextConfig = {
//   reactStrictMode: true,
//   // Убираем standalone - он только усложняет
// }
//
// module.exports = nextConfig

// next.config.js
/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,

  // Отключаем статическую генерацию для страниц с динамическими данными
  experimental: {
    isrMemoryCacheSize: 0,
  },

  // Исключаем проблемные страницы из статической генерации
  exportPathMap: async function (
      defaultPathMap,
      { dev, dir, outDir, distDir, buildId }
  ) {
    const pathMap = { ...defaultPathMap };
    // Удаляем проблемные маршруты из статической генерации
    delete pathMap['/resumes'];
    return pathMap;
  }
};

module.exports = nextConfig;
