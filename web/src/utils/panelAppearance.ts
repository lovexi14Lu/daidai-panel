export interface PanelAppearanceSettings {
  log_background_color?: string
  log_background_image?: string
}

const DEFAULT_LOG_BACKGROUND_COLOR = '#0f172a'

function toCSSImageValue(image?: string) {
  const trimmed = image?.trim() || ''
  if (!trimmed) {
    return 'none'
  }

  return `url("${trimmed.replace(/"/g, '\\"')}")`
}

export function applyPanelAppearance(settings?: PanelAppearanceSettings | null) {
  const root = document.documentElement
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
