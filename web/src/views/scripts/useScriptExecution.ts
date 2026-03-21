import { onBeforeUnmount, ref, watch, type Ref } from 'vue'
import { ElMessage } from 'element-plus'
import { scriptApi } from '@/api/script'

interface UseScriptExecutionOptions {
  selectedFile: Ref<string>
  fileContent: Ref<string>
}

export function useScriptExecution({ selectedFile, fileContent }: UseScriptExecutionOptions) {
  const showDebugDialog = ref(false)
  const debugCode = ref('')
  const debugFileName = ref('')
  const debugRunId = ref('')
  const debugLogs = ref<string[]>([])
  const debugRunning = ref(false)
  const debugError = ref('')
  const debugExitCode = ref<number | null>(null)
  const debugCodeChanged = ref(false)

  const showCodeRunner = ref(false)
  const runnerCode = ref('')
  const runnerLanguage = ref('python')
  const runnerRunId = ref('')
  const runnerLogs = ref<string[]>([])
  const runnerRunning = ref(false)
  const runnerExitCode = ref<number | null>(null)

  let debugTimer: ReturnType<typeof setInterval> | null = null
  let runnerTimer: ReturnType<typeof setInterval> | null = null

  function getFileName(path: string) {
    return path.split('/').pop() || path
  }

  watch(showDebugDialog, (val) => {
    if (!val && debugTimer) {
      clearInterval(debugTimer)
      debugTimer = null
      debugRunning.value = false
    }
  })

  watch(showCodeRunner, (val) => {
    if (!val && runnerTimer) {
      clearInterval(runnerTimer)
      runnerTimer = null
      runnerRunning.value = false
    }
  })

  onBeforeUnmount(() => {
    if (debugTimer) {
      clearInterval(debugTimer)
      debugTimer = null
    }
    if (runnerTimer) {
      clearInterval(runnerTimer)
      runnerTimer = null
    }
  })

  async function handleDebugRun() {
    if (!selectedFile.value) return
    debugCode.value = fileContent.value
    debugFileName.value = getFileName(selectedFile.value)
    debugLogs.value = []
    debugRunning.value = false
    debugError.value = ''
    debugExitCode.value = null
    debugRunId.value = ''
    debugCodeChanged.value = false
    showDebugDialog.value = true
  }

  async function handleDebugStart() {
    if (!selectedFile.value) return
    if (debugCodeChanged.value) {
      try {
        await scriptApi.saveContent(selectedFile.value, debugCode.value)
        fileContent.value = debugCode.value
        debugCodeChanged.value = false
      } catch {
        ElMessage.error('保存代码失败')
        return
      }
    }
    debugLogs.value = []
    debugError.value = ''
    debugExitCode.value = null
    debugRunning.value = true
    try {
      const res = await scriptApi.debugRun({ path: selectedFile.value })
      debugRunId.value = res.run_id
      pollDebugLogs()
    } catch (err: any) {
      debugError.value = err?.response?.data?.error || err?.message || '调试运行失败'
      ElMessage.error(debugError.value)
      debugRunning.value = false
    }
  }

  function pollDebugLogs() {
    if (debugTimer) clearInterval(debugTimer)
    debugTimer = setInterval(async () => {
      if (!debugRunId.value) {
        if (debugTimer) {
          clearInterval(debugTimer)
          debugTimer = null
        }
        return
      }
      try {
        const res = await scriptApi.debugLogs(debugRunId.value)
        debugLogs.value = res.data.logs || []
        if (res.data.done) {
          debugRunning.value = false
          if (debugTimer) {
            clearInterval(debugTimer)
            debugTimer = null
          }
          if (res.data.status === 'failed') {
            debugExitCode.value = res.data.exit_code ?? null
            debugError.value = 'failed'
          }
        }
      } catch {
        debugRunning.value = false
        if (debugTimer) {
          clearInterval(debugTimer)
          debugTimer = null
        }
      }
    }, 500)
  }

  async function handleDebugStop() {
    if (!debugRunId.value) return
    try {
      await scriptApi.debugStop(debugRunId.value)
    } catch {
      // ignore
    }
    debugRunning.value = false
    if (debugTimer) {
      clearInterval(debugTimer)
      debugTimer = null
    }
    try {
      const res = await scriptApi.debugLogs(debugRunId.value)
      debugLogs.value = res.data.logs || []
    } catch {
      // ignore
    }
  }

  function openCodeRunner() {
    runnerCode.value = ''
    runnerLanguage.value = 'python'
    runnerLogs.value = []
    runnerRunning.value = false
    runnerExitCode.value = null
    runnerRunId.value = ''
    showCodeRunner.value = true
  }

  async function handleRunCode() {
    if (!runnerCode.value.trim()) {
      ElMessage.warning('请输入代码')
      return
    }
    runnerLogs.value = []
    runnerExitCode.value = null
    runnerRunning.value = true
    try {
      const res = await scriptApi.runCode(runnerCode.value, runnerLanguage.value)
      runnerRunId.value = res.run_id
      pollRunnerLogs()
    } catch (err: any) {
      ElMessage.error(err?.response?.data?.error || '运行失败')
      runnerRunning.value = false
    }
  }

  function pollRunnerLogs() {
    if (runnerTimer) clearInterval(runnerTimer)
    runnerTimer = setInterval(async () => {
      if (!runnerRunId.value) {
        if (runnerTimer) {
          clearInterval(runnerTimer)
          runnerTimer = null
        }
        return
      }
      try {
        const res = await scriptApi.debugLogs(runnerRunId.value)
        runnerLogs.value = res.data.logs || []
        if (res.data.done) {
          runnerRunning.value = false
          runnerExitCode.value = res.data.exit_code ?? null
          if (runnerTimer) {
            clearInterval(runnerTimer)
            runnerTimer = null
          }
        }
      } catch {
        runnerRunning.value = false
        if (runnerTimer) {
          clearInterval(runnerTimer)
          runnerTimer = null
        }
      }
    }, 500)
  }

  async function handleStopRunner() {
    if (!runnerRunId.value) return
    try {
      await scriptApi.debugStop(runnerRunId.value)
    } catch {
      // ignore
    }
    runnerRunning.value = false
    if (runnerTimer) {
      clearInterval(runnerTimer)
      runnerTimer = null
    }
  }

  return {
    showDebugDialog,
    debugCode,
    debugFileName,
    debugLogs,
    debugRunning,
    debugError,
    debugExitCode,
    debugCodeChanged,
    showCodeRunner,
    runnerCode,
    runnerLanguage,
    runnerLogs,
    runnerRunning,
    runnerExitCode,
    handleDebugRun,
    handleDebugStart,
    handleDebugStop,
    openCodeRunner,
    handleRunCode,
    handleStopRunner
  }
}
