/**
 * 权限工具函数
 * Permission utility functions
 */

/**
 * 检查用户是否拥有指定角色
 * Check if user has specified role
 * @param {string[]} userRoles - 用户角色列表 User roles list
 * @param {string|string[]} requiredRoles - 所需角色 Required roles
 * @returns {boolean} - 是否拥有权限 Whether has permission
 */
export const hasRole = (userRoles, requiredRoles) => {
  if (!userRoles || !requiredRoles) return false
  
  // 如果requiredRoles是字符串，转换为数组
  // If requiredRoles is string, convert to array
  const roles = Array.isArray(requiredRoles) ? requiredRoles : [requiredRoles]
  
  // 检查用户是否拥有任一所需角色
  // Check if user has any of the required roles
  return userRoles.some(role => roles.includes(role))
}

/**
 * 检查用户是否拥有指定权限
 * Check if user has specified permission
 * @param {string[]} userPermissions - 用户权限列表 User permissions list
 * @param {string|string[]} requiredPermissions - 所需权限 Required permissions
 * @returns {boolean} - 是否拥有权限 Whether has permission
 */
export const hasPermission = (userPermissions, requiredPermissions) => {
  if (!userPermissions || !requiredPermissions) return false
  
  // 如果requiredPermissions是字符串，转换为数组
  // If requiredPermissions is string, convert to array
  const permissions = Array.isArray(requiredPermissions) 
    ? requiredPermissions 
    : [requiredPermissions]
  
  // 检查用户是否拥有所有所需权限
  // Check if user has all required permissions
  return permissions.every(permission => userPermissions.includes(permission))
} 