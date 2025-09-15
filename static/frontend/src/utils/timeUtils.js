/**
 * 时间工具类
 * 提供各种时间格式转换功能
 */

/**
 * 将ISO格式的时间字符串转换为更易读的格式
 * @param {string} isoString - ISO格式的时间字符串，如：2025-09-14T23:32:08.508+08:00
 * @param {string} format - 输出格式，默认为 'YYYY-MM-DD HH:mm:ss'
 * @returns {string} 格式化后的时间字符串
 */
export function formatISODate(isoString, format = 'YYYY-MM-DD HH:mm:ss') {
  if (!isoString) return '';
  
  try {
    const date = new Date(isoString);
    
    // 检查日期是否有效
    if (isNaN(date.getTime())) {
      console.error('Invalid date:', isoString);
      return '';
    }
    
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    const seconds = String(date.getSeconds()).padStart(2, '0');
    
    // 替换格式中的占位符
    return format
      .replace('YYYY', year)
      .replace('MM', month)
      .replace('DD', day)
      .replace('HH', hours)
      .replace('mm', minutes)
      .replace('ss', seconds);
  } catch (error) {
    console.error('Error formatting date:', error);
    return '';
  }
}

/**
 * 获取相对时间描述，如"刚刚"、"5分钟前"、"2小时前"等
 * @param {string} isoString - ISO格式的时间字符串
 * @returns {string} 相对时间描述
 */
export function getRelativeTime(isoString) {
  if (!isoString) return '';
  
  try {
    const date = new Date(isoString);
    const now = new Date();
    const diffInSeconds = Math.floor((now - date) / 1000);
    
    if (diffInSeconds < 0) {
      return formatISODate(isoString);
    }
    
    if (diffInSeconds < 60) {
      return '刚刚';
    }
    
    if (diffInSeconds < 3600) {
      return `${Math.floor(diffInSeconds / 60)}分钟前`;
    }
    
    if (diffInSeconds < 86400) {
      return `${Math.floor(diffInSeconds / 3600)}小时前`;
    }
    
    if (diffInSeconds < 2592000) {
      return `${Math.floor(diffInSeconds / 86400)}天前`;
    }
    
    if (diffInSeconds < 31536000) {
      return `${Math.floor(diffInSeconds / 2592000)}个月前`;
    }
    
    return `${Math.floor(diffInSeconds / 31536000)}年前`;
  } catch (error) {
    console.error('Error getting relative time:', error);
    return '';
  }
}

/**
 * 格式化日期为指定格式
 * @param {Date|string|number} date - 日期对象、时间戳或ISO字符串
 * @param {string} format - 输出格式，默认为 'YYYY-MM-DD'
 * @returns {string} 格式化后的日期字符串
 */
export function formatDate(date, format = 'YYYY-MM-DD') {
  if (!date) return '';
  
  try {
    // 转换输入为Date对象
    const dateObj = date instanceof Date ? date : new Date(date);
    
    // 检查日期是否有效
    if (isNaN(dateObj.getTime())) {
      console.error('Invalid date:', date);
      return '';
    }
    
    const year = dateObj.getFullYear();
    const month = String(dateObj.getMonth() + 1).padStart(2, '0');
    const day = String(dateObj.getDate()).padStart(2, '0');
    const hours = String(dateObj.getHours()).padStart(2, '0');
    const minutes = String(dateObj.getMinutes()).padStart(2, '0');
    const seconds = String(dateObj.getSeconds()).padStart(2, '0');
    const milliseconds = String(dateObj.getMilliseconds()).padStart(3, '0');
    const weekday = ['日', '一', '二', '三', '四', '五', '六'][dateObj.getDay()];
    
    return format
      .replace('YYYY', year)
      .replace('YY', String(year).slice(2))
      .replace('MM', month)
      .replace('M', String(dateObj.getMonth() + 1))
      .replace('DD', day)
      .replace('D', String(dateObj.getDate()))
      .replace('HH', hours)
      .replace('H', String(dateObj.getHours()))
      .replace('mm', minutes)
      .replace('m', String(dateObj.getMinutes()))
      .replace('ss', seconds)
      .replace('s', String(dateObj.getSeconds()))
      .replace('SSS', milliseconds)
      .replace('E', weekday);
  } catch (error) {
    console.error('Error formatting date:', error);
    return '';
  }
}

/**
 * 获取当前时间的ISO字符串
 * @returns {string} 当前时间的ISO字符串
 */
export function getCurrentISOString() {
  return new Date().toISOString();
}

/**
 * 获取当前时间的格式化字符串
 * @param {string} format - 输出格式，默认为 'YYYY-MM-DD HH:mm:ss'
 * @returns {string} 格式化后的当前时间字符串
 */
export function getCurrentFormattedDate(format = 'YYYY-MM-DD HH:mm:ss') {
  return formatDate(new Date(), format);
}

/**
 * 计算两个日期之间的差值
 * @param {Date|string|number} date1 - 第一个日期
 * @param {Date|string|number} date2 - 第二个日期
 * @param {string} unit - 返回差值的单位，可选值：'milliseconds', 'seconds', 'minutes', 'hours', 'days'
 * @returns {number} 两个日期之间的差值
 */
export function getDateDiff(date1, date2, unit = 'days') {
  if (!date1 || !date2) return 0;
  
  try {
    const dateObj1 = date1 instanceof Date ? date1 : new Date(date1);
    const dateObj2 = date2 instanceof Date ? date2 : new Date(date2);
    
    // 检查日期是否有效
    if (isNaN(dateObj1.getTime()) || isNaN(dateObj2.getTime())) {
      console.error('Invalid date:', date1, date2);
      return 0;
    }
    
    const diffInMilliseconds = Math.abs(dateObj2 - dateObj1);
    
    switch (unit) {
      case 'milliseconds':
        return diffInMilliseconds;
      case 'seconds':
        return Math.floor(diffInMilliseconds / 1000);
      case 'minutes':
        return Math.floor(diffInMilliseconds / (1000 * 60));
      case 'hours':
        return Math.floor(diffInMilliseconds / (1000 * 60 * 60));
      case 'days':
        return Math.floor(diffInMilliseconds / (1000 * 60 * 60 * 24));
      default:
        return diffInMilliseconds;
    }
  } catch (error) {
    console.error('Error calculating date difference:', error);
    return 0;
  }
}

/**
 * 判断日期是否为今天
 * @param {Date|string|number} date - 要判断的日期
 * @returns {boolean} 是否为今天
 */
export function isToday(date) {
  if (!date) return false;
  
  try {
    const dateObj = date instanceof Date ? date : new Date(date);
    const today = new Date();
    
    return (
      dateObj.getDate() === today.getDate() &&
      dateObj.getMonth() === today.getMonth() &&
      dateObj.getFullYear() === today.getFullYear()
    );
  } catch (error) {
    console.error('Error checking if date is today:', error);
    return false;
  }
}

/**
 * 解析各种格式的日期字符串为Date对象
 * @param {string} dateString - 日期字符串
 * @returns {Date|null} 解析后的Date对象，解析失败则返回null
 */
export function parseDate(dateString) {
  if (!dateString) return null;
  
  try {
    // 尝试直接解析
    const date = new Date(dateString);
    if (!isNaN(date.getTime())) {
      return date;
    }
    
    // 尝试解析常见的日期格式
    const formats = [
      // ISO格式: 2025-09-14T23:32:08.508+08:00
      /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?$/,
      // 标准日期格式: 2025-09-14 23:32:08
      /^(\d{4})-(\d{2})-(\d{2}) (\d{2}):(\d{2}):(\d{2})$/,
      // 日期格式: 2025-09-14
      /^(\d{4})-(\d{2})-(\d{2})$/,
      // 日期格式: 2025/09/14
      /^(\d{4})\/(\d{2})\/(\d{2})$/,
      // 日期格式: 14/09/2025
      /^(\d{2})\/(\d{2})\/(\d{4})$/,
      // 日期格式: 14-09-2025
      /^(\d{2})-(\d{2})-(\d{4})$/
    ];
    
    for (const format of formats) {
      const match = dateString.match(format);
      if (match) {
        // 根据不同格式解析日期
        if (format === formats[0] || format === formats[1]) {
          // ISO格式或标准日期格式
          return new Date(dateString);
        } else if (format === formats[2] || format === formats[3]) {
          // YYYY-MM-DD 或 YYYY/MM/DD
          const [, year, month, day] = match;
          return new Date(parseInt(year), parseInt(month) - 1, parseInt(day));
        } else {
          // DD/MM/YYYY 或 DD-MM-YYYY
          const [, day, month, year] = match;
          return new Date(parseInt(year), parseInt(month) - 1, parseInt(day));
        }
      }
    }
    
    return null;
  } catch (error) {
    console.error('Error parsing date:', error);
    return null;
  }
}