module.exports = {
  // 生成器配置
  generatorName: "javascript",
  // 输出目录
  outputDir: "src/api",
  // 生成器额外配置
  additionalProperties: {
    usePromises: true,
    modelPropertyNaming: "camelCase",
    apiPackage: "api",
    modelPackage: "models"
  },
  // 全局配置
  globalProperties: {
    apiDocs: "http://localhost:8080/swagger/doc.json",
    // 支持 Promise
    supportPromise: true,
    // 移除前缀
    removeOperationIdPrefix: true,
  },
} 