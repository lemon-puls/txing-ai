-- 简化测试工作流：专门测试条件节点的各种判断类型
-- 包含：表达式判断、AI判断、工具结果判断

INSERT INTO `agent_flows` (`name`, `description`, `topology`, `create_time`, `update_time`)
VALUES (
    '条件判断测试-简化版',
    '专门测试条件节点功能，包含表达式判断、AI判断、错误处理策略',
    '{
  "nodes": [
    {
      "id": "start_1",
      "type": "start",
      "position": { "x": 50, "y": 200 },
      "data": {
        "label": "开始",
        "nodeType": "start"
      }
    },
    {
      "id": "llm_input",
      "type": "llm",
      "position": { "x": 250, "y": 200 },
      "data": {
        "label": "分析输入",
        "nodeType": "llm",
        "modelConfig": {
          "model": "",
          "systemPrompt": "请分析用户输入的内容类型，简单输出：\\n- 如果是问题，输出: type=question\\n- 如果是任务，输出: type=task\\n- 如果是闲聊，输出: type=chat\\n\\n只输出这一行，不要其他内容。",
          "temperature": 0.1,
          "maxTokens": 100,
          "contextEnabled": false
        }
      }
    },
    {
      "id": "condition_type",
      "type": "condition",
      "position": { "x": 500, "y": 200 },
      "data": {
        "label": "输入类型判断",
        "nodeType": "condition",
        "conditionConfig": {
          "type": "expression",
          "expression": "{{output}} contains \\'task\\'",
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
      "id": "llm_task",
      "type": "llm",
      "position": { "x": 800, "y": 100 },
      "data": {
        "label": "任务处理",
        "nodeType": "llm",
        "modelConfig": {
          "model": "",
          "systemPrompt": "你是一个任务处理助手。请帮用户处理他们的任务：\\n1. 分析任务需求\\n2. 提供解决方案\\n3. 给出执行步骤",
          "temperature": 0.5,
          "maxTokens": 1500,
          "contextEnabled": true
        }
      }
    },
    {
      "id": "llm_chat",
      "type": "llm",
      "position": { "x": 800, "y": 350 },
      "data": {
        "label": "闲聊回复",
        "nodeType": "llm",
        "modelConfig": {
          "model": "",
          "systemPrompt": "你是一个友好的聊天助手。请用轻松愉快的语气回复用户，可以适当使用表情。",
          "temperature": 0.9,
          "maxTokens": 500,
          "contextEnabled": true
        }
      }
    },
    {
      "id": "condition_sentiment",
      "type": "condition",
      "position": { "x": 1100, "y": 100 },
      "data": {
        "label": "情感判断",
        "nodeType": "condition",
        "conditionConfig": {
          "type": "llm",
          "expression": "",
          "llmPrompt": "请判断以下内容的情感倾向：\\n- 如果是正面、积极的，返回 true\\n- 如果是负面、消极的，返回 false\\n\\n严格按照 JSON 格式返回：{\\"result\\": true/false, \\"reason\\": \\"判断原因\\"}",
          "toolName": "",
          "toolResultKey": "",
          "expectedValue": "",
          "failureAction": "default_false",
          "failureBranch": "false"
        }
      }
    },
    {
      "id": "llm_positive",
      "type": "llm",
      "position": { "x": 1400, "y": 50 },
      "data": {
        "label": "正面强化",
        "nodeType": "llm",
        "modelConfig": {
          "model": "",
          "systemPrompt": "用户表达了积极的情感。请用更加积极正面的方式回应，鼓励用户继续努力！",
          "temperature": 0.7,
          "maxTokens": 300,
          "contextEnabled": true
        }
      }
    },
    {
      "id": "llm_comfort",
      "type": "llm",
      "position": { "x": 1400, "y": 200 },
      "data": {
        "label": "安慰鼓励",
        "nodeType": "llm",
        "modelConfig": {
          "model": "",
          "systemPrompt": "用户似乎有些消极情绪。请用温暖、理解的语气回应，给予安慰和鼓励。记住：每个人都会有低谷期，重要的是保持希望。",
          "temperature": 0.7,
          "maxTokens": 500,
          "contextEnabled": true
        }
      }
    },
    {
      "id": "end_success",
      "type": "end",
      "position": { "x": 1700, "y": 125 },
      "data": {
        "label": "完成",
        "nodeType": "end"
      }
    }
  ],
  "edges": [
    {
      "id": "edge_1",
      "source": "start_1",
      "target": "llm_input",
      "sourceHandle": "output",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_2",
      "source": "llm_input",
      "target": "condition_type",
      "sourceHandle": "output",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_3",
      "source": "condition_type",
      "target": "llm_task",
      "sourceHandle": "true",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_4",
      "source": "condition_type",
      "target": "llm_chat",
      "sourceHandle": "false",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_5",
      "source": "llm_task",
      "target": "condition_sentiment",
      "sourceHandle": "output",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_6",
      "source": "llm_chat",
      "target": "end_success",
      "sourceHandle": "output",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_7",
      "source": "condition_sentiment",
      "target": "llm_positive",
      "sourceHandle": "true",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_8",
      "source": "condition_sentiment",
      "target": "llm_comfort",
      "sourceHandle": "false",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_9",
      "source": "llm_positive",
      "target": "end_success",
      "sourceHandle": "output",
      "targetHandle": "input",
      "type": "smoothstep",
      "animated": true
    },
    {
      "id": "edge_10",
      "source": "llm_comfort",
      "target": "end_success",
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
