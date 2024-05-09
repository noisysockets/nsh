import globals from "globals";
import pluginJs from "@eslint/js";
import tseslint from "typescript-eslint";
import pluginReactConfig from "eslint-plugin-react/configs/recommended.js";


export default [
  {ignores: ["node_modules", "dist"]},
  {languageOptions: { globals: globals.browser }},
  pluginJs.configs.recommended,
  ...tseslint.configs.recommended,
  {...pluginReactConfig, settings: {react: { version: "detect" } }, rules: {"react/react-in-jsx-scope": 0}},
];