import { defineConfig } from "eslint/config";
import globals from "globals";
import js from "@eslint/js";
import pluginVue from "eslint-plugin-vue";
import pluginJs from "@eslint/js";
import eslintPluginPrettierRecommended from "eslint-plugin-prettier/recommended";

export default defineConfig([
  {
    ignores: ["node_modules/", "dist/"],
  },
  { files: ["**/*.{js,mjs,cjs,vue}"] },
  {
    files: ["**/*.{js,mjs,cjs,vue}"],
    languageOptions: { globals: globals.browser },
  },
  {
    files: ["**/*.{js,mjs,cjs,vue}"],
    plugins: { js },
    extends: ["js/recommended"],
  },
  pluginVue.configs["flat/essential"],
  pluginJs.configs.recommended,
  eslintPluginPrettierRecommended,
]);
