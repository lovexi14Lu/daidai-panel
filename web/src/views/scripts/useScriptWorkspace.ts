import { onBeforeUnmount, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useScriptWorkspaceActions } from './useScriptWorkspaceActions'
import { useScriptWorkspaceBrowser } from './useScriptWorkspaceBrowser'

export function useScriptWorkspace() {
  const router = useRouter()
  const route = useRoute()

  const browser = useScriptWorkspaceBrowser()
  const actions = useScriptWorkspaceActions({
    selectedFile: browser.selectedFile,
    fileContent: browser.fileContent,
    originalContent: browser.originalContent,
    isBinary: browser.isBinary,
    isEditing: browser.isEditing,
    hasChanges: browser.hasChanges,
    loadTree: browser.loadTree,
    loadFileContent: browser.loadFileContent
  })

  onMounted(() => {
    window.addEventListener('keydown', actions.handleKeyDown)
    window.addEventListener('resize', browser.handleResize)
  })

  async function openFileFromRoute(fileParam?: string) {
    if (!fileParam) return

    const previousSelectedFile = browser.selectedFile.value
    browser.selectedFile.value = fileParam
    const loaded = await browser.loadFileContent(fileParam)
    if (!loaded) {
      browser.selectedFile.value = previousSelectedFile
    } else {
      browser.mobileShowEditor.value = true
    }
    await router.replace({ path: '/scripts' })
  }

  watch(
    () => route.query.file,
    (fileParam) => {
      if (typeof fileParam !== 'string' || !fileParam.trim()) {
        return
      }
      void openFileFromRoute(fileParam)
    },
    { immediate: true }
  )

  onBeforeUnmount(() => {
    window.removeEventListener('keydown', actions.handleKeyDown)
    window.removeEventListener('resize', browser.handleResize)
  })

  void browser.loadTree()

  return {
    ...browser,
    ...actions
  }
}
