import path from 'node:path'
import { defineConfig, loadEnv } from 'vite'
import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { viteEnvs } from 'vite-envs'

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    viteEnvs({
      declarationFile: '.env.declaration',
      computedEnv: ({ resolvedConfig }) => {
        const env = loadEnv(resolvedConfig.mode, resolvedConfig.root, 'WEB_')
        return {
          WEB_API_URL: env.WEB_API_URL || 'http://localhost:8080',
        }
      },
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
