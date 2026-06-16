import { defineConfig } from 'vite'
import * as path from 'path'
import * as dotenv from 'dotenv'
import * as fs from 'fs'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

const NODE_ENV = process.env.NODE_ENV || 'development'
const envFile = `.env.${NODE_ENV}`
const envConfig = dotenv.parse(fs.readFileSync(envFile))
for (const k in envConfig) {
  process.env[k] = envConfig[k]
}

let alias = {
  '@': path.resolve(__dirname, './src'),
  'vue$': 'vue/dist/vue.runtime.esm-bundler.js',
}

const conf = {
  base: './', // The location of the index.html file
  root: './', // Resource path imported by js, src
  server: {
    open: true,
    port: process.env.VITE_DEV_PORT,
    proxy: {
      [process.env.VITE_SERVER_API]: {
        target: process.env.VITE_SERVER_PATH,
        // rewrite: path => path.replace(/^\/api/, '/api'), //For simulation
        changeOrigin: true,
      },
    },
  },
  build: {
    target: 'es2020',
    minify: 'esbuild', // Whether to compress, boolean | 'terser' | 'esbuild', esbuild is used by default
    manifest: false, // Whether to output maifest.json
    sourcemap: false, // Whether to output soucemap.json
    emptyOutDir: true,
    outDir: 'dist', // output directory
    rollupOptions: {
      output: {
        manualChunks (id) {
          if (id.includes('node_modules')) {
            const arr = id.toString().split('node_modules/')[1].split('/')
            switch (arr[0]) {
              case '@popperjs':
              case '@vue':
              case 'axios':
              case 'element-plus':
              case '@element-plus':
                return '_' + arr[0]
              default :
                return '__vendor'
            }
          }else if(id.includes('Gwen-admin/src')){
            //Everything under src is packaged together, otherwise there will be many small files.
            return 'gwen'
          }
        },
        chunkFileNames: 'static/chunk/[name]-[hash].js',
        entryFileNames: 'static/entry/[name]-[hash].js',
        assetFileNames: 'static/[ext]/[name]-[hash].[ext]'
      },
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        javascriptEnabled: true,
      },
    },
  },
  resolve: {
    alias,
  },
  plugins: [
    tailwindcss(),
    vue(),
  ],
}

// https://vitejs.dev/config/
export default defineConfig(conf)
