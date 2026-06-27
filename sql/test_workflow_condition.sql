-- 测试工作流：智能研究助手（包含所有节点类型）
-- 包含：开始、LLM、工具、条件分支（表达式/AI判断）、结束

INSERT INTO `agent_flows` (`name`, `description`, `topology`, `create_time`, `update_time`)
VALUES (
    '智能研究助手-测试工作流',
    '测试工作流：包含所有节点类型，演示条件分支、AI判断、表达式判断等功能',
    '{
  "nodes": [
    {
      "id": "start_1",
      "type": "start",
      "position": { "x": 50, "y": 250 },
      "data": {
        "label": "开始",
        "nodeType": "start"
      }
    },
    {
      "id": "llm_analyze",
      "type": "llm",
      "position": { "x": 250, "y": 250 },
      "data": {
        "label": "需求分析",
        "nodeType": "llm",
        "modelConfig": {
          "model": "",
          "systemPrompt": "你是一个专业的研究助手。请分析用户的输入，提取关键信息：\\n1. 研究主题\\n2. 关键词列表\\n3. 研究方向\\n\\n请用以下格式输出：\\n主题: [研究主题]\\n关键词: [关键词1, 关键词2, ...]\\n方向: [研究方向描述]",
          "temperature": 0.3,
          "maxTokens": 1024,
          "contextEnabled": false
        }
      }
    },
    {
      "id": "tool_search",
      "type": "tool",
      "position": { "x": 500, "y": 250 },
      "data": {
        "label": "搜索信息",
        "nodeType": "tool",
        "toolConfig": {
          "tools": ["web_search_tool"],
          "params": {}
        }
      }
    },
    {
      "id": "condition_search",
      "type": "condition",
      "position": { "x": 750, "y": 250 },
      "data": {
        "label": "搜索结果判断",
        "nodeType": "condition",
        "conditionConfig": {
          "type": "expression",
          "expression": "{{output}} not_equals ''",
          "llmPrompt": "",
          "toolName": "",
          "toolResultKey": "",
          "expectedValue": "",
          "failureAction": "default_false",
          "failureBranch": "false"
        }
      }
    },
    {
      "id": "llm_generate",
      "type": "llm",
      "position": { "x": 1050, "y": 150 },
      "data": {
        "label": "内容生成",
        "nodeType": "llm",
        "modelConfig": {
          "model": "",
          "systemPrompt": "你是一个专业的内容创作者。根据用户的研究主题和搜索到的信息，生成一篇高质量的研究报告。\\n\\n要求：\\n1. 结构清晰，包含摘要、正文、结论\\n2. 内容准确，引用可靠来源\\n3. 字数 500-1000 字\\n4. 使用 Markdown 格式",
          "temperature": 0.7,
          "maxTokens": 2048,
          "contextEnabled": true
        }
      }
    },
    {
      "id": "condition_quality",
      "type": "condition",
      "position": { "x": 1300, "y": 150 },
      "data": {
        "label": "质量判断",
        "nodeType": "condition",
        "conditionConfig": {
          "type": "llm",
          "expression": "",
          "llmPrompt": "请评估以下内容的质量，判断是否满足以下标准：\\n1. 结构完整（有摘要、正文、结论）\\n2. 内容充实（字数不少于 500 字）\\n3. 逻辑清晰\\n\\n如果满足所有标准，返回 true；否则返回 false。",
          "toolName": "",
          "toolResultKey": "",
          "expectedValue": "",
          "failureAction": "default_false",
          "failureBranch": "false"
        }
      }
    },
    {
      "id": "tool_save",
      "type": "tool",
      "position": { "x": 1550, "y": 100 },
      "data": {
        "label": "保存文档",
        "nodeType": "tool",
        "toolConfig": {
          "tools": ["markdown_save_tool"],
          "params": {}
        }
      }
    },
    {
      "id": "llm_error",
      "type": "llm",
      "position": { "x": 1050, "y": 400 },
      "data": {
        "label": "错误提示",
        "nodeType": "llm",
        "modelConfig": {
          "model": "",
          "systemPrompt": "你是一个友好的助手。用户的研究任务遇到了问题，请用简洁友好的语言说明情况，并建议用户：\\n1. 检查网络连接\\n2. 尝试更具体的搜索词\\n3. 稍后重试\\n\\n保持语气积极正面。",
          "temperature": 0.5,
          "maxTokens": 512,
          "contextEnabled": false
        }
      }
    },
    {
      "id": "llm_retry",
      "type": "llm",
      "position": { "x": 1550, "y": 300 },
      "data": {
        "label": "优化提示",
        "nodeType": "llm",
        "modelConfig": {
          "model": "",
          "systemPrompt": "你是一个内容优化助手。之前的生成内容质量不达标，请根据原始需求重新生成一篇更高质量的研究报告。\\n\\n优化方向：\\n1. 增加内容深度\\n2. 完善文章结构\\n3. 补充更多细节和例子\\n4. 确保字数不少于 500 字",
          "temperature": 0.6,
          "maxTokens": 2048,
          "contextEnabled": true
        }
      }
    },
    {
      "id": "end_success",
      "type": "end",
      "position": { "x": 1800, "y": 100 },
      "data": {
        "label": "完成",
        "nodeType": "end"
      }
    },
    {
      "id": "end_error",
      "type": "end",
      "position": { "x": 1300, "y": 400 },
      "data": {
        "label": "结束（错误）",
        "nodeType": "end"
      }
    }
  ],
  "edges": [
    {
      "id": "edge_1",
      "source": "start_1",
      "target": "llm_analyze",
      "sourceHandle": "output",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_2",
      "source": "llm_analyze",
      "target": "tool_search",
      "sourceHandle": "output",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_3",
      "source": "tool_search",
      "target": "condition_search",
      "sourceHandle": "output",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_4",
      "source": "condition_search",
      "target": "llm_generate",
      "sourceHandle": "true",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_5",
      "source": "condition_search",
      "target": "llm_error",
      "sourceHandle": "false",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_6",
      "source": "llm_generate",
      "target": "condition_quality",
      "sourceHandle": "output",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_7",
      "source": "condition_quality",
      "target": "tool_save",
      "sourceHandle": "true",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_8",
      "source": "condition_quality",
      "target": "llm_retry",
      "sourceHandle": "false",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_9",
      "source": "tool_save",
      "target": "end_success",
      "sourceHandle": "output",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_10",
      "source": "llm_error",
      "target": "end_error",
      "sourceHandle": "output",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_11",
      "source": "llm_retry",
      "target": "condition_quality",
      "sourceHandle": "output",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    }
  ]
}',
    NOW(),
    NOW()
);
