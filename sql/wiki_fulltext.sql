-- LLM Wiki 知识库 FULLTEXT 迁移
-- GORM AutoMigrate 不支持 FULLTEXT INDEX，需在部署后手动执行一次（重复执行会报 Duplicate key name，可忽略）。
-- 依赖 MySQL 5.7.6+ 的 ngram 解析器（中文全文检索）；若服务端 ngram_token_size 需调整，需在 my.cnf 配置并重启。

USE {{DATABASE}}; -- 部署时替换为实际库名

ALTER TABLE wiki_pages
    ADD FULLTEXT INDEX ft_wiki_pages_title_content (title, content) WITH PARSER ngram;
