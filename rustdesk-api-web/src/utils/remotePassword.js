import { ElMessage, ElMessageBox } from 'element-plus'
import { T } from '@/utils/i18n'

function escapeHtml (text) {
  return String(text)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

function decodeBase64Utf8 (value) {
  const binary = atob(value)
  const bytes = Uint8Array.from(binary, (c) => c.charCodeAt(0))
  return new TextDecoder().decode(bytes)
}

export function structuredPasswordFromId (deviceId) {
  if (!deviceId) {
    return ''
  }
  const id = String(deviceId).trim()
  if (!id) {
    return ''
  }
  const tail = id.length >= 5 ? id.slice(-5) : id.padStart(5, '0')
  return `Rd@${tail}`
}

export function decodeRemotePassword (hashOrPassword, deviceId = '') {
  if (!hashOrPassword) {
    return structuredPasswordFromId(deviceId)
  }
  const raw = String(hashOrPassword).trim()
  if (!raw) {
    return structuredPasswordFromId(deviceId)
  }

  try {
    const decoded = decodeBase64Utf8(raw)
    if (decoded && /^[\x20-\x7E]+$/.test(decoded)) {
      return decoded
    }
  } catch (_) {
    // not base64
  }

  try {
    const decoded = atob(raw)
    if (decoded && /^[\x20-\x7E]+$/.test(decoded)) {
      return decoded
    }
  } catch (_) {
    // not base64
  }

  if (/^[\x20-\x7E]+$/.test(raw) && !/^[A-Za-z0-9+/=]{8,}$/.test(raw)) {
    return raw
  }

  return structuredPasswordFromId(deviceId)
}

async function copyText (text) {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(T('CopySuccess') || 'Sao chep thanh cong')
  } catch (_) {
    ElMessage.error(T('CopyFailed') || 'Sao chep that bai')
  }
}

export async function showRemotePasswordDialog (hashOrPassword, meta = {}) {
  const password = decodeRemotePassword(hashOrPassword, meta.id)
  if (!password) {
    ElMessage.warning(T('NoRemotePassword') || 'Khong co mat khau remote cho thiet bi nay.')
    return
  }

  const label = [meta.id, meta.hostname || meta.alias].filter(Boolean).join(' / ')
  const title = (T('RemotePassword') || 'Mat khau remote') + (label ? ` — ${label}` : '')
  const hint = T('DecodedRemotePassword') || 'Mat khau da giai ma'
  const body = `
    <div style="text-align:center;padding:4px 0 8px">
      <div style="color:#909399;font-size:13px;margin-bottom:12px">${escapeHtml(hint)}</div>
      <code style="display:block;font-size:24px;font-weight:600;letter-spacing:1px;padding:16px 12px;background:#f5f7fa;border-radius:8px;color:#303133;font-family:Consolas,Monaco,monospace">${escapeHtml(password)}</code>
    </div>
  `

  try {
    await ElMessageBox.confirm(body, title, {
      confirmButtonText: T('Copy') || 'Sao chep',
      cancelButtonText: T('Close') || 'Dong',
      type: 'info',
      distinguishCancelAndClose: true,
      dangerouslyUseHTMLString: true,
    })
    await copyText(password)
  } catch (_) {
    // closed
  }
}
