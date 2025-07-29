# frontend

This template should help get you started developing with Vue 3 in Vite.

## Recommended IDE Setup

[VSCode](https://code.visualstudio.com/) + [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur).

## Customize configuration

See [Vite Configuration Reference](https://vite.dev/config/).

## Project Setup

```sh
npm install
```

### Compile and Hot-Reload for Development

```sh
npm run dev
```

### Compile and Minify for Production

```sh
npm run build
```

### Lint with [ESLint](https://eslint.org/)

```sh
npm run lint
```

### 生成 api 代码

```sh
npm run generate-api
```

## 环境变量配置

本项目使用环境变量进行配置。请在项目根目录创建 `.env` 文件，并添加以下配置：

```
# 分析脚本配置
VITE_ANALYTICS_SCRIPT_URL=your_analytics_script_url
VITE_ANALYTICS_WEBSITE_ID=your_website_id
```

在 GitHub Actions 中，这些环境变量通过 GitHub Secrets 提供。