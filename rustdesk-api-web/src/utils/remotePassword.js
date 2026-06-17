import { ElMessage, ElMessageBox } from 'element-plus'
import { T } from '@/utils/i18n'
import { handleClipboard } from '@/utils/clipboard'

export function decodeRemotePassword (hashOrPassword) {
  if (!hashOrPassword) {
    return ''
  }
  const raw = String(hashOrPassword).trim()
  if (!raw) {
    return ''
  }
  try {
    const decoded = atob(raw)
    if (/^[\x20-\x7E]+$/.test(decoded)) {
      return decoded
    }
  } catch (_) {
    // not base64
  }
  if (/^[\x20-\x7E]+$/.test(raw)) {
    return raw
  }
  return ''
}

export async function showRemotePasswordDialog (hashOrPassword, meta = {}) {
  const password = decodeRemotePassword(hashOrPassword)
  if (!password) {
    ElMessage.warning(T('NoRemotePassword') || 'Khong co mat khau remote cho thiet bi nay.')
    return
  }
  const label = [meta.id, meta.hostname || meta.alias].filter(Boolean).join(' / ')
  const title = (T('RemotePassword') || 'Mat khau remote') + (label ? ` — ${label}` : '')
  try {
    await ElMessageBox.confirm(password, title, {
      confirmButtonText: T('Copy') || 'Sao chep',
      cancelButtonText: T('Close') || 'Dong',
      type: 'info',
      distinguishCancelAndClose: true,
    })
    handleClipboard(password)
  } catch (_) {
    // closed
  }
}
