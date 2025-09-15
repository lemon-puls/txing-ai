/**
 * timeUtils 使用示例
 */
import {
  formatISODate,
  getRelativeTime,
  formatDate,
  getCurrentISOString,
  getCurrentFormattedDate,
  getDateDiff,
  isToday,
  parseDate
} from './timeUtils';

// 示例ISO时间字符串
const isoDateString = '2025-09-14T23:32:08.508+08:00';

// 示例1: 将ISO格式转换为易读格式
console.log('示例1 - 格式化ISO日期:');
console.log(`原始ISO日期: ${isoDateString}`);
console.log(`默认格式化: ${formatISODate(isoDateString)}`);
console.log(`自定义格式化: ${formatISODate(isoDateString, 'YYYY年MM月DD日 HH时mm分ss秒')}`);
console.log('-------------------');

// 示例2: 获取相对时间
console.log('示例2 - 相对时间:');
// 注意: 这里使用当前时间减去一定时间来模拟过去的时间点
const now = new Date();
const fiveMinutesAgo = new Date(now.getTime() - 5 * 60 * 1000).toISOString();
const twoHoursAgo = new Date(now.getTime() - 2 * 60 * 60 * 1000).toISOString();
const threeDaysAgo = new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000).toISOString();

console.log(`5分钟前: ${getRelativeTime(fiveMinutesAgo)}`);
console.log(`2小时前: ${getRelativeTime(twoHoursAgo)}`);
console.log(`3天前: ${getRelativeTime(threeDaysAgo)}`);
console.log('-------------------');

// 示例3: 格式化日期
console.log('示例3 - 格式化日期:');
const dateObj = new Date('2025-09-14');
console.log(`默认格式 (YYYY-MM-DD): ${formatDate(dateObj)}`);
console.log(`带时间格式 (YYYY-MM-DD HH:mm:ss): ${formatDate(dateObj, 'YYYY-MM-DD HH:mm:ss')}`);
console.log(`中文格式 (YYYY年MM月DD日 星期E): ${formatDate(dateObj, 'YYYY年MM月DD日 星期E')}`);
console.log('-------------------');

// 示例4: 获取当前时间
console.log('示例4 - 当前时间:');
console.log(`当前ISO时间: ${getCurrentISOString()}`);
console.log(`当前格式化时间: ${getCurrentFormattedDate()}`);
console.log(`当前日期: ${getCurrentFormattedDate('YYYY-MM-DD')}`);
console.log('-------------------');

// 示例5: 计算日期差值
console.log('示例5 - 日期差值:');
const date1 = new Date('2025-09-14');
const date2 = new Date('2025-10-24');
console.log(`两个日期相差天数: ${getDateDiff(date1, date2, 'days')}天`);
console.log(`两个日期相差小时数: ${getDateDiff(date1, date2, 'hours')}小时`);
console.log('-------------------');

// 示例6: 判断是否为今天
console.log('示例6 - 是否为今天:');
console.log(`今天的日期是否为今天: ${isToday(new Date())}`);
console.log(`2025-09-14是否为今天: ${isToday('2025-09-14')}`);
console.log('-------------------');

// 示例7: 解析各种格式的日期
console.log('示例7 - 解析日期:');
const dateFormats = [
  '2025-09-14T23:32:08.508+08:00',
  '2025-09-14 23:32:08',
  '2025-09-14',
  '2025/09/14',
  '14/09/2025',
  '14-09-2025'
];

dateFormats.forEach(dateStr => {
  const parsedDate = parseDate(dateStr);
  console.log(`解析 "${dateStr}": ${parsedDate ? formatDate(parsedDate, 'YYYY-MM-DD HH:mm:ss') : '解析失败'}`); 
});