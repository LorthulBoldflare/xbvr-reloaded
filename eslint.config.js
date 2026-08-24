// ESLint 9 flat config (the eslintrc block in package.json was dead config,
// and referenced the uninstalled babel-eslint parser).
const pluginVue = require('eslint-plugin-vue')
const babelParser = require('@babel/eslint-parser')

module.exports = [
  ...pluginVue.configs['flat/essential'],
  {
    files: ['ui/src/**/*.{js,vue}'],
    languageOptions: {
      parserOptions: {
        parser: babelParser,
        sourceType: 'module',
        ecmaVersion: 'latest',
        requireConfigFile: false,
        babelOptions: { configFile: './ui/babel.config.js' }
      }
    },
    rules: {}
  },
  {
    ignores: ['ui/dist/**', 'node_modules/**']
  }
]
