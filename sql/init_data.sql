CREATE TABLE `agent_flows` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `create_time` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `update_time` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `delete_time` datetime(3) DEFAULT NULL COMMENT '删除时间,为NULL则未删除',
  `name` varchar(100) NOT NULL COMMENT '工作流名称',
  `description` varchar(500) DEFAULT NULL COMMENT '工作流描述',
  `topology` text DEFAULT NULL COMMENT '工作流拓扑图数据',
  PRIMARY KEY (`id`),
  KEY `idx_agent_flows_delete_time` (`delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入示例工作流：智能问答助手
-- 流程：开始 -> 理解问题(LLM) -> 搜索信息(工具) -> 生成答案(LLM) -> 结束
INSERT INTO `agent_flows` (`name`, `description`, `topology`, `create_time`, `update_time`) VALUES
('智能问答助手', '一个经典的AI问答工作流，包含问题理解、信息搜索和答案生成三个步骤',
'{
  "nodes": [
    {
      "id": "node_start_001",
      "type": "start",
      "position": {"x": 100, "y": 200},
      "data": {
        "label": "开始",
        "nodeType": "start"
      }
    },
    {
      "id": "node_llm_understand",
      "type": "llm",
      "position": {"x": 320, "y": 200},
      "data": {
        "label": "理解问题",
        "nodeType": "llm",
        "description": "分析用户输入，提取关键信息",
        "modelConfig": {
          "model": "deepseek-chat",
          "systemPrompt": "你是一个专业的问答助手。请仔细分析用户的问题，提取其中的关键词和核心意图。用简洁的语言总结问题的关键点。",
          "temperature": 0.3,
          "maxTokens": 1024,
          "contextEnabled": true
        }
      }
    },
    {
      "id": "node_tool_search",
      "type": "tool",
      "position": {"x": 540, "y": 200},
      "data": {
        "label": "搜索信息",
        "nodeType": "tool",
        "description": "通过搜索工具获取相关信息",
        "toolConfig": {
          "tools": ["web_search_tool"],
          "params": {}
        }
      }
    },
    {
      "id": "node_llm_generate",
      "type": "llm",
      "position": {"x": 760, "y": 200},
      "data": {
        "label": "生成答案",
        "nodeType": "llm",
        "description": "基于搜索结果生成完整答案",
        "modelConfig": {
          "model": "deepseek-chat",
          "systemPrompt": "你是一个专业的问答助手。请根据用户的问题和搜索到的信息，生成一个完整、准确、有帮助的回答。回答要条理清晰，语言简洁。",
          "temperature": 0.7,
          "maxTokens": 2048,
          "contextEnabled": true
        }
      }
    },
    {
      "id": "node_end_001",
      "type": "end",
      "position": {"x": 980, "y": 200},
      "data": {
        "label": "结束",
        "nodeType": "end"
      }
    }
  ],
  "edges": [
    {
      "id": "edge_001",
      "source": "node_start_001",
      "target": "node_llm_understand",
      "sourceHandle": "output",
      "targetHandle": "input"
    },
    {
      "id": "edge_002",
      "source": "node_llm_understand",
      "target": "node_tool_search",
      "sourceHandle": "output",
      "targetHandle": "input"
    },
    {
      "id": "edge_003",
      "source": "node_tool_search",
      "target": "node_llm_generate",
      "sourceHandle": "output",
      "targetHandle": "input"
    },
    {
      "id": "edge_004",
      "source": "node_llm_generate",
      "target": "node_end_001",
      "sourceHandle": "output",
      "targetHandle": "input"
    }
  ]
}', NOW(3), NOW(3));');
