import { onBeforeUnmount, onMounted } from 'vue'
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

    const fileParam = route.query.file as string
    if (fileParam) {
      browser.selectedFile.value = fileParam
      void browser.loadFileContent(fileParam)
      browser.mobileShowEditor.value = true
      void router.replace({ path: '/scripts' })
    }
  })

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
