// 通用 OSS 上传工具函数
// Reusable OSS upload helper (used by ImageUploader, MediaUploader, etc.)
// 流程：1) 获取 upload presigned URL  2) PUT 文件到该 URL  3) 获取 download presigned URL 返回前端
// Flow: 1) get upload presigned URL  2) PUT file to it  3) get download presigned URL for the response

import { defaultApi } from '@/api'

/**
 * @typedef {Object} UploadResult
 * @property {string} url       下载访问 URL（带签名，前端可直接展示）
 * @property {string} key       对象 key（仅文件名）
 * @property {string} mimeType  上传时的 Content-Type
 */

/**
 * 把一个 File/Blob 上传到 OSS
 * Upload a File/Blob to OSS via presigned URL
 *
 * @param {File|Blob} file 待上传文件
 * @param {Object} [options]
 * @param {string} [options.key] 自定义对象 key；缺省自动生成时间戳随机数
 * @param {string} [options.mimeType] 自定义 Content-Type；缺省从 file.type 推断
 * @param {string} [options.keyPrefix] key 前缀（用于目录归类，如 "media/"），不含尾斜杠
 * @param {string} [options.ext] 显式指定扩展名（覆盖自动推断）
 * @returns {Promise<UploadResult>}
 */
export async function uploadFileToOSS(file, options = {}) {
  let { key, mimeType, keyPrefix = '', ext } = options

  if (!mimeType) {
    mimeType = file.type || 'application/octet-stream'
  }

  if (!ext) {
    if (file.name && file.name.includes('.')) {
      ext = file.name.split('.').pop()
    } else if (mimeType.startsWith('image/')) {
      ext = mimeType.split('/')[1] || 'png'
    } else if (mimeType.startsWith('video/')) {
      ext = mimeType.split('/')[1] || 'mp4'
    } else {
      ext = 'bin'
    }
  }

  if (!key) {
    key = `${keyPrefix}${keyPrefix ? '/' : ''}${Date.now()}-${Math.floor(Math.random() * 100000)}.${ext}`
  }

  // 1) 获取上传预签名 URL
  // Get presigned URL for upload
  const upRes = await defaultApi.apiCosPresignedUrlPost({
    type: 'upload',
    key
  })
  if (upRes?.code !== 0) {
    throw new Error(upRes?.msg || '获取上传地址失败')
  }

  // 2) PUT 文件到 OSS
  // PUT file to OSS
  const putResp = await fetch(upRes.data.url, {
    method: 'PUT',
    body: file,
    headers: { 'Content-Type': mimeType }
  })
  if (!putResp.ok) {
    throw new Error(`上传失败（HTTP ${putResp.status}）`)
  }

  // 3) 获取下载预签名 URL
  // Get presigned URL for download
  const dlRes = await defaultApi.apiCosPresignedUrlPost({
    type: 'download',
    key
  })
  if (dlRes?.code !== 0) {
    throw new Error(dlRes?.msg || '获取访问地址失败')
  }

  return {
    url: dlRes.data.url,
    key,
    mimeType
  }
}

/**
 * 压缩图片到 maxWidth/maxHeight 内，返回 Blob
 * Compress an image (File/Blob) and return a JPEG Blob.
 * @param {File|Blob} file
 * @param {Object} [opts] { quality=0.8, maxWidth=1920, maxHeight=1920, mime='image/jpeg' }
 * @returns {Promise<Blob>}
 */
export function compressImage(file, opts = {}) {
  const { quality = 0.8, maxWidth = 1920, maxHeight = 1920, mime = 'image/jpeg' } = opts
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = (e) => {
      const img = new Image()
      img.src = e.target.result
      img.onload = () => {
        let { width, height } = img
        if (width > maxWidth || height > maxHeight) {
          const ratio = Math.min(maxWidth / width, maxHeight / height)
          width = Math.floor(width * ratio)
          height = Math.floor(height * ratio)
        }
        const canvas = document.createElement('canvas')
        canvas.width = width
        canvas.height = height
        const ctx = canvas.getContext('2d')
        ctx.fillStyle = '#fff'
        ctx.fillRect(0, 0, canvas.width, canvas.height)
        ctx.drawImage(img, 0, 0, width, height)
        canvas.toBlob((blob) => {
          if (blob) resolve(blob)
          else reject(new Error('图片压缩失败'))
        }, mime, quality)
      }
      img.onerror = () => reject(new Error('图片解码失败'))
    }
    reader.onerror = () => reject(new Error('文件读取失败'))
  })
}

/**
 * 判断文件是否为图片
 * Check whether a file is an image.
 * @param {File|Blob} file
 */
export function isImage(file) {
  return !!file && (file.type || '').startsWith('image/')
}

/**
 * 判断文件是否为视频
 * Check whether a file is a video.
 * @param {File|Blob} file
 */
export function isVideo(file) {
  return !!file && (file.type || '').startsWith('video/')
}