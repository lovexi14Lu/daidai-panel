const SUBSCRIPTION_LABEL_PREFIX = 'subscription:'
const SUBSCRIPTION_DISPLAY_LABEL = '订阅任务'

function uniqueLabels(labels: string[]) {
  return Array.from(new Set(labels.filter(Boolean)))
}

export function isInternalTaskLabel(label: string) {
  return label.startsWith(SUBSCRIPTION_LABEL_PREFIX)
}

export function getDisplayTaskLabels(labels: string[] = []) {
  const displayLabels: string[] = []
  let hasSubscriptionLabel = false

  for (const label of labels) {
    if (!label) continue
    if (isInternalTaskLabel(label)) {
      hasSubscriptionLabel = true
      continue
    }
    displayLabels.push(label)
  }

  if (hasSubscriptionLabel) {
    displayLabels.push(SUBSCRIPTION_DISPLAY_LABEL)
  }

  return uniqueLabels(displayLabels)
}

export function splitTaskLabels(labels: string[] = []) {
  const editableLabels: string[] = []
  const internalLabels: string[] = []

  for (const label of labels) {
    if (!label) continue
    if (isInternalTaskLabel(label)) {
      internalLabels.push(label)
      continue
    }
    editableLabels.push(label)
  }

  return {
    editableLabels: uniqueLabels(editableLabels),
    internalLabels: uniqueLabels(internalLabels),
  }
}

export function mergeTaskLabels(editableLabels: string[] = [], internalLabels: string[] = []) {
  return uniqueLabels([...editableLabels, ...internalLabels])
}
