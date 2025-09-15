# 时间工具类使用文档

## 简介

`timeUtils.js` 提供了一系列处理日期和时间的实用函数，可以帮助您在项目中轻松处理各种时间格式转换、计算和展示需求。

## 安装

该工具类已经集成在项目中，无需额外安装。

## 使用方法

### 导入

```javascript
import { formatISODate, getRelativeTime, formatDate /* 其他需要的函数 */ } from '@/utils/timeUtils';
```

### 主要功能

#### 1. 格式化ISO日期字符串

将ISO格式的时间字符串转换为更易读的格式。

```javascript
import { formatISODate } from '@/utils/timeUtils';

// 默认格式 (YYYY-MM-DD HH:mm:ss)
formatISODate('2025-09-14T23:32:08.508+08:00');
// 输出: "2025-09-14 23:32:08"

// 自定义格式
formatISODate('2025-09-14T23:32:08.508+08:00', 'YYYY年MM月DD日 HH时mm分ss秒');
// 输出: "2025年09月14日 23时32分08秒"
```

#### 2. 获取相对时间描述

将时间转换为"刚刚"、"5分钟前"、"2小时前"等相对时间描述。

```javascript
import { getRelativeTime } from '@/utils/timeUtils';

// 假设当前时间是 2025-09-14 23:40:00
getRelativeTime('2025-09-14T23:32:08.508+08:00');
// 输出: "8分钟前"

getRelativeTime('2025-09-14T21:32:08.508+08:00');
// 输出: "2小时前"

getRelativeTime('2025-09-12T23:32:08.508+08:00');
// 输出: "2天前"
```

#### 3. 格式化日期

格式化日期为指定格式。

```javascript
import { formatDate } from '@/utils/timeUtils';

// 默认格式 (YYYY-MM-DD)
formatDate(new Date('2025-09-14'));
// 输出: "2025-09-14"

// 自定义格式
formatDate('2025-09-14', 'YYYY年MM月DD日 星期E');
// 输出: "2025年09月14日 星期日"
```

#### 4. 获取当前时间

```javascript
import { getCurrentISOString, getCurrentFormattedDate } from '@/utils/timeUtils';

// 获取当前时间的ISO字符串
getCurrentISOString();
// 输出: "2025-09-14T23:32:08.508Z"

// 获取当前时间的格式化字符串
getCurrentFormattedDate();
// 输出: "2025-09-14 23:32:08"

// 自定义格式
getCurrentFormattedDate('YYYY年MM月DD日');
// 输出: "2025年09月14日"
```

#### 5. 计算日期差值

```javascript
import { getDateDiff } from '@/utils/timeUtils';

const date1 = new Date('2025-09-14');
const date2 = new Date('2025-10-24');

// 计算天数差
getDateDiff(date1, date2, 'days');
// 输出: 40

// 计算小时差
getDateDiff(date1, date2, 'hours');
// 输出: 960
```

#### 6. 判断日期是否为今天

```javascript
import { isToday } from '@/utils/timeUtils';

isToday(new Date());
// 输出: true

isToday('2025-09-14'); // 假设今天是2025-09-14
// 输出: true

isToday('2025-09-15'); // 假设今天是2025-09-14
// 输出: false
```

#### 7. 解析各种格式的日期字符串

```javascript
import { parseDate, formatDate } from '@/utils/timeUtils';

const date = parseDate('2025-09-14T23:32:08.508+08:00');
// 返回Date对象

formatDate(date, 'YYYY-MM-DD HH:mm:ss');
// 输出: "2025-09-14 23:32:08"

// 支持多种格式
parseDate('2025-09-14');
parseDate('2025/09/14');
parseDate('14/09/2025');
parseDate('14-09-2025');
```

## 格式化占位符

在格式化日期时，可以使用以下占位符：

- `YYYY`: 四位年份，如 2025
- `YY`: 两位年份，如 25
- `MM`: 两位月份，如 09
- `M`: 一位月份，如 9
- `DD`: 两位日期，如 14
- `D`: 一位日期，如 14
- `HH`: 两位小时，如 23
- `H`: 一位小时，如 23
- `mm`: 两位分钟，如 32
- `m`: 一位分钟，如 32
- `ss`: 两位秒数，如 08
- `s`: 一位秒数，如 8
- `SSS`: 三位毫秒数，如 508
- `E`: 星期几，如 日、一、二、三、四、五、六

## 示例

更多使用示例请参考 `timeUtils.example.js` 文件。

## 单元测试

所有函数都有对应的单元测试，请参考 `timeUtils.test.js` 文件。