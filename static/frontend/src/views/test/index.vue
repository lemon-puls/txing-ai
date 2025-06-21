<template>
  <div>
    <!-- 聊天消息区域 -->
    <div class="chat-messages custom-scrollbar" ref="messagesContainer">
      <TransitionGroup name="message">
        <div
          v-for="message in currentChat.messages"
          :key="message.id"
          class="message-item"
          :class="message.role"
        >
          <div class="message-avatar">
            <el-avatar
              :size="40"
              :src="message.role === 'user' ? userAvatar : (currentChat.assistant?.avatar || aiAvatar)"
            >
              {{ message.role === 'user' ? 'U' : (currentChat.assistant?.name?.charAt(0) || 'AI') }}
            </el-avatar>
          </div>
          <div class="message-content">
            <div class="message-text" v-html="renderMessage(message.content)"></div>
            <div class="message-actions">
              <el-button-group>
                <el-button text size="small" @click="copyMessage(message)">
                  <template #icon>
                    <CopyDocument/>
                  </template>
                  复制
                </el-button>
                <el-button text size="small" @click="regenerateMessage(message)" v-if="message.role === 'assistant'">
                  <template #icon>
                    <RefreshRight/>
                  </template>
                  重新生成
                </el-button>
              </el-button-group>
            </div>
          </div>
        </div>
      </TransitionGroup>
    </div>
  </div>
</template>

<script setup name="ChatPage">
import {ref, onMounted} from 'vue'
import {CopyDocument, RefreshRight} from '@element-plus/icons-vue'
import {ElMessage} from 'element-plus'
import {marked} from 'marked';
import hljs from 'highlight.js';
import 'highlight.js/styles/atom-one-dark.css';
// 需在marked配置前注册Java语言
import java from 'highlight.js/lib/languages/java';

hljs.registerLanguage('java', java);

marked.setOptions({
  highlight: (code, lang) => {
    const language = hljs.getLanguage(lang) ? lang : 'plaintext';
    return hljs.highlight(code, {language}).value;
  },
  langPrefix: 'hljs language-', // highlight.js class prefix
  gfm: true, // 启用 GitHub 风格的 Markdown
  breaks: true // 启用换行符转换
});
// 渲染消息内容
const renderMessage = (content) => {
  try {
    const rendered = marked(content);
    return rendered;
  } catch (err) {
    console.error('Markdown rendering error:', err);
    return content;
  }
}

// 状态
const currentChat = ref({
  id: 1,
  title: '冒泡排序实现',
  model: 'gpt-3.5-turbo',
  webSearch: false,
  maxTokens: 2048,
  temperature: 1,
  topP: 0.7,
  topK: 50,
  presencePenalty: 0,
  frequencyPenalty: 0,
  repetitionPenalty: 1,
  messages: [
    {
      id: 1,
      role: 'assistant',
      content: `# Java实现冒泡排序

冒泡排序是一种简单的排序算法，它重复地遍历要排序的列表，比较相邻的元素并交换它们的位置，直到列表排序完成。

以下是Java实现冒泡排序的代码：

\`\`\`java
public class BubbleSort {

    public static void bubbleSort(int[] arr) {
        int n = arr.length;
        // 外层循环控制排序轮数
        for (int i = 0; i < n - 1; i++) {
            // 内层循环控制每轮比较次数
            for (int j = 0; j < n - i - 1; j++) {
                // 如果前一个元素比后一个元素大，则交换它们
                if (arr[j] > arr[j + 1]) {
                    // 交换arr[j]和arr[j+1]
                    int temp = arr[j];
                    arr[j] = arr[j + 1];
                    arr[j + 1] = temp;
                }
            }
        }
    }

    // 优化版的冒泡排序（如果某一轮没有发生交换，说明已经有序）
    public static void optimizedBubbleSort(int[] arr) {
        int n = arr.length;
        boolean swapped;
        for (int i = 0; i < n - 1; i++) {
            swapped = false;
            for (int j = 0; j < n - i - 1; j++) {
                if (arr[j] > arr[j + 1]) {
                    // 交换arr[j]和arr[j+1]
                    int temp = arr[j];
                    arr[j] = arr[j + 1];
                    arr[j + 1] = temp;
                    swapped = true;
                }
            }
            // 如果没有发生交换，提前结束排序
            if (!swapped) {
                break;
            }
        }
    }

    public static void main(String[] args) {
        int[] arr = {64, 34, 25, 12, 22, 11, 90};

        System.out.println("排序前的数组:");
        printArray(arr);

        bubbleSort(arr);
        // 或者使用优化版: optimizedBubbleSort(arr);

        System.out.println("排序后的数组:");
        printArray(arr);
    }

    // 辅助方法：打印数组
    public static void printArray(int[] arr) {
        for (int value : arr) {
            System.out.print(value + " ");
        }
        System.out.println();
    }
}
\`\`\`

## 代码说明

1. **基本冒泡排序**：
   - 外层循环控制排序轮数（n-1轮）
   - 内层循环比较相邻元素，如果顺序不对就交换
   - 每轮结束后，最大的元素会"冒泡"到数组末尾

2. **优化版冒泡排序**：
   - 添加了\`swapped\`标志位
   - 如果某一轮没有发生交换，说明数组已经有序，可以提前结束排序
   - 对于基本有序的数组，能显著提高效率

3. **时间复杂度**：
   - 最坏情况：O(n²)（完全逆序）
   - 最好情况：O(n)（已经有序，使用优化版）
   - 平均情况：O(n²)

4. **空间复杂度**：O(1)，是原地排序算法

5. **稳定性**：冒泡排序是稳定的排序算法，因为相等的元素不会交换位置

你可以根据需要选择基本版或优化版的实现。对于小型数组或基本有序的数组，冒泡排序是一个不错的选择。`
    }
  ]
})
// 头像
const userAvatar = 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png'
const aiAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

const copyMessage = async (message) => {
  try {
    await navigator.clipboard.writeText(message.content)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

onMounted(() => {
  hljs.highlightAll()
})
</script>

<style scoped lang="scss">
.message-content {
  .message-text {
    pre {
      margin: 1em 0;
      padding: 0;
      background: transparent;

      code.hljs {
        display: block;
        overflow-x: auto;
        padding: 1em;
        background: #282c34;
        border-radius: 4px;
        font-family: 'Fira Code', monospace;
        font-size: 14px;
        line-height: 1.5;
      }
    }
  }
}
</style>
