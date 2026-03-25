export interface PanelAppearanceSettings {
  editor_background_color?: string
  log_background_color?: string
  log_background_image?: string
}

const DEFAULT_LOG_BACKGROUND_COLOR = '#0f172a'
const DEFAULT_EDITOR_BACKGROUND_COLOR = '#111827'
const DEFAULT_EDITOR_FOREGROUND_COLOR = '#e5e7eb'

function toCSSImageValue(image?: string) {
  const trimmed = image?.trim() || ''
  if (!trimmed) {
    return 'none'
  }

  return `url("${trimmed.replace(/"/g, '\\"')}")`
}

function parseColor(color?: string) {
  const text = color?.trim() || ''
  if (!text) return null

  if (text.startsWith('#')) {
    const hex = text.slice(1)
    if (hex.length === 3) {
      const r = Number.parseInt(hex.charAt(0) + hex.charAt(0), 16)
      const g = Number.parseInt(hex.charAt(1) + hex.charAt(1), 16)
      const b = Number.parseInt(hex.charAt(2) + hex.charAt(2), 16)
      return Number.isNaN(r) || Number.isNaN(g) || Number.isNaN(b) ? null : { r, g, b }
    }
    if (hex.length === 6 || hex.length === 8) {
      const offset = hex.length === 8 ? 2 : 0
      const r = Number.parseInt(hex.slice(offset, offset + 2), 16)
      const g = Number.parseInt(hex.slice(offset + 2, offset + 4), 16)
      const b = Number.parseInt(hex.slice(offset + 4, offset + 6), 16)
      return Number.isNaN(r) || Number.isNaN(g) || Number.isNaN(b) ? null : { r, g, b }
    }
  }

  const match = text.match(/^rgba?\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})(?:\s*,\s*[0-9.]+\s*)?\)$/i)
  if (!match) {
    return null
  }

  const r = Number.parseInt(match[1] ?? '', 10)
  const g = Number.parseInt(match[2] ?? '', 10)
  const b = Number.parseInt(match[3] ?? '', 10)
  return Number.isNaN(r) || Number.isNaN(g) || Number.isNaN(b) ? null : { r, g, b }
}

function getReadableTextColor(background?: string) {
  const rgb = parseColor(background)
  if (!rgb) {
    return DEFAULT_EDITOR_FOREGROUND_COLOR
  }

  const toLinear = (channel: number) => {
    const value = channel / 255
    return value <= 0.03928 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4
  }

  const luminance = 0.2126 * toLinear(rgb.r) + 0.7152 * toLinear(rgb.g) + 0.0722 * toLinear(rgb.b)
  return luminance < 0.45 ? DEFAULT_EDITOR_FOREGROUND_COLOR : '#111827'
}

export function applyPanelAppearance(settings?: PanelAppearanceSettings | null) {
  const root = document.documentElement
  const editorBackground = settings?.editor_background_color?.trim() || DEFAULT_EDITOR_BACKGROUND_COLOR
  root.style.setProperty('--dd-editor-bg-color', editorBackground)
  root.style.setProperty('--dd-editor-fg-color', getReadableTextColor(editorBackground))
  root.style.setProperty('--dd-log-bg-color', settings?.log_background_color?.trim() || DEFAULT_LOG_BACKGROUND_COLOR)
  root.style.setProperty('--dd-log-bg-image', toCSSImageValue(settings?.log_background_image))
}

export async function fetchAndApplyPanelAppearance() {
  try {
    const response = await fetch('/api/system/panel-settings', { cache: 'no-store' })
    if (!response.ok) {
      return
    }

    const payload = await response.json() as { data?: PanelAppearanceSettings }
    applyPanelAppearance(payload.data || null)
  } catch {
    // ignore startup appearance load failures
  }
}
