import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";
import obfuscator from "vite-plugin-bundle-obfuscator";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    obfuscator({
      excludes: [],
      enable: true,
      log: true,
      autoExcludeNodeModules: true,
      threadPool: true,
    }),
  ],
});
