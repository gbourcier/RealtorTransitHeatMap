import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vuetify, { transformAssetUrls } from 'vite-plugin-vuetify'
import ViteFonts from 'unplugin-fonts/vite'

export default defineConfig({
  plugins: [
    vue({ template: { transformAssetUrls } }),
    vuetify({
      autoImport: true,
      styles: { configFile: 'src/styles/settings.scss' },
    }),
    ViteFonts({
      google: {
        families: [
          { name: 'Roboto', styles: 'wght@100;300;400;500;700;900' },
        ],
      },
    }),
  ],
  define: { 'process.env': {} },
  css: {
    preprocessorOptions: {
      scss: { api: 'modern-compiler' },
      sass: { api: 'modern-compiler' },
    },
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
    extensions: ['.js', '.json', '.jsx', '.mjs', '.ts', '.tsx', '.vue'],
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          vue: ['vue', 'vue-router', 'pinia'],
          vuetify: ['vuetify'],
          map: ['leaflet', 'leaflet.markercluster', 'h3-js'],
          http: ['axios'],
        },
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:3000',
        changeOrigin: true,
      },
    },
  },
})
