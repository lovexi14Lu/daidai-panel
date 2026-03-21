import { ref } from 'vue'
import { systemApi } from '@/api/system'
import { ElLoading, ElMessage, ElMessageBox } from 'element-plus'

export function useSettingsOverview() {
  const systemInfo = ref<any>({})
  const systemStats = ref<any>(null)
  const currentVersion = ref('')
  const updateInfo = ref<any>(null)
  const updateStatus = ref<any>(null)
  const checkingUpdate = ref(false)
  const updatingPanel = ref(false)

  function formatBytes(bytes: number): string {
    if (!bytes) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
  }

  function getUsageClass(percent: number): string {
    if (!percent) return ''
    if (percent < 60) return 'usage-success'
    if (percent < 80) return 'usage-warning'
    return 'usage-danger'
  }

  async function loadSystemInfo() {
    try {
      const res = await systemApi.info()
      systemInfo.value = res.data || {}
    } catch {
      // ignore
    }
  }

  async function loadSystemStats() {
    try {
      const res = await systemApi.stats()
      systemStats.value = res.data || {}
    } catch {
      // ignore
    }
  }

  async function loadVersion() {
    try {
      const res = await systemApi.version()
      currentVersion.value = res.data.version || ''
    } catch {
      // ignore
    }
  }

  async function handleCheckUpdate() {
    checkingUpdate.value = true
    try {
      const res = await systemApi.checkUpdate()
      updateInfo.value = res.data
      if (res.data.has_update) {
        if (res.data.auto_update_supported) {
          ElMessage.success(`发现新版本 v${res.data.latest}，可在下方直接点击“立即更新”`)
        } else {
          ElMessage.warning(res.data.update_disabled_reason || '当前部署暂不支持面板内一键更新')
        }
      } else {
        ElMessage.success(`当前版本 v${res.data.current} 已经是最新版了`)
      }
    } catch (err: any) {
      const msg = err?.response?.data?.error || '检查更新失败，请稍后重试'
      ElMessage.error(msg)
    } finally {
      checkingUpdate.value = false
    }
  }

  async function handleUpdatePanel() {
    if (updatingPanel.value) {
      ElMessage.warning('更新任务已经在进行中，请稍候')
      return
    }

    if (updateInfo.value?.auto_update_supported === false) {
      ElMessage.warning(updateInfo.value?.update_disabled_reason || '当前部署暂不支持面板内一键更新')
      return
    }

    try {
      await ElMessageBox.confirm(
        '确认开始更新面板吗？系统会先拉取最新镜像，再自动重建容器。更新期间服务会短暂中断。',
        '立即更新',
        {
          confirmButtonText: '开始更新',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
    } catch (err: any) {
      if (err === 'cancel' || err?.toString?.() === 'cancel') {
        return
      }
      ElMessage.error(err?.message || '无法确认更新操作')
      return
    }

    const loading = ElLoading.service({
      lock: true,
      text: '正在提交更新任务，请稍候...',
      background: 'rgba(0, 0, 0, 0.7)'
    }) as any

    updatingPanel.value = true
    updateStatus.value = {
      status: 'running',
      phase: 'preparing',
      message: '正在提交更新任务'
    }

    try {
      const res = await systemApi.updatePanel()
      updateStatus.value = res.data || updateStatus.value
      syncLoadingText(loading, updateStatus.value?.message)
      await waitForUpdateResult(loading)
    } catch (err: any) {
      loading.close()
      updatingPanel.value = false
      const msg = err?.response?.data?.error || err?.message || '更新失败，请手动更新'
      ElMessage.error(msg)
    }
  }

  async function waitForUpdateResult(loading: any) {
    let sawProgress = false
    let lastStatus = updateStatus.value?.status || ''
    const startedAt = Date.now()

    while (Date.now() - startedAt < 12 * 60 * 1000) {
      await delay(2000)

      try {
        const res = await systemApi.updateStatus()
        updateStatus.value = res.data || {}
        lastStatus = updateStatus.value?.status || lastStatus
        sawProgress = true
        syncLoadingText(loading, updateStatus.value?.message)

        if (lastStatus === 'failed') {
          throw new Error(updateStatus.value?.error || updateStatus.value?.message || '更新失败')
        }

        if (lastStatus === 'restarting') {
          loading.close()
          await waitForAvailability('面板正在切换到新版本，请稍候...')
          return
        }
      } catch (err: any) {
        if (shouldTreatAsRestart(err, sawProgress, lastStatus)) {
          loading.close()
          await waitForAvailability('面板正在切换到新版本，请稍候...')
          return
        }
        throw err
      }
    }

    loading.close()
    updatingPanel.value = false
    ElMessage.warning('更新超时，请手动刷新页面检查')
  }

  async function handleRestartPanel() {
    try {
      await ElMessageBox.confirm('确定要重启面板吗？重启期间服务将短暂中断。', '重启面板', {
        confirmButtonText: '确认重启',
        cancelButtonText: '取消',
        type: 'warning'
      })
      await systemApi.restart()
      waitForRestart()
    } catch {
      // cancelled
    }
  }

  function waitForRestart() {
    const loading = ElLoading.service({
      lock: true,
      text: '面板正在重启，请稍候...',
      background: 'rgba(0, 0, 0, 0.7)'
    })
    let attempts = 0
    setTimeout(() => {
      const poll = setInterval(async () => {
        attempts++
        try {
          const res = await fetch('/', { method: 'HEAD' })
          if (res.ok) {
            clearInterval(poll)
            loading.close()
            window.location.reload()
          }
        } catch {
          // ignore
        }
        if (attempts >= 60) {
          clearInterval(poll)
          loading.close()
          ElMessage.warning('重启超时，请手动刷新页面')
        }
      }, 2000)
    }, 3000)
  }

  async function waitForAvailability(text: string) {
    const loading = ElLoading.service({
      lock: true,
      text,
      background: 'rgba(0, 0, 0, 0.7)'
    })
    let attempts = 0

    setTimeout(() => {
      const poll = setInterval(async () => {
        attempts++
        try {
          const res = await fetch('/', { method: 'HEAD', cache: 'no-store' })
          if (res.ok) {
            clearInterval(poll)
            loading.close()
            window.location.reload()
          }
        } catch {
          // ignore
        }

        if (attempts >= 80) {
          clearInterval(poll)
          loading.close()
          updatingPanel.value = false
          ElMessage.warning('等待新版本启动超时，请手动刷新页面检查')
        }
      }, 3000)
    }, 2000)
  }

  function shouldTreatAsRestart(err: any, sawProgress: boolean, lastStatus: string) {
    if (!sawProgress) {
      return false
    }

    if (err?.response) {
      return false
    }

    return lastStatus === 'running' || lastStatus === 'restarting'
  }

  function syncLoadingText(loading: any, text?: string) {
    if (!text || !loading || typeof loading.setText !== 'function') {
      return
    }
    loading.setText(text)
  }

  function delay(ms: number) {
    return new Promise((resolve) => setTimeout(resolve, ms))
  }

  function openGitHub() {
    const url = updateInfo.value?.has_update && updateInfo.value?.release_url
      ? updateInfo.value.release_url
      : 'https://github.com/linzixuanzz/daidai-panel/releases'
    window.open(url, '_blank')
  }

  return {
    systemInfo,
    systemStats,
    currentVersion,
    updateInfo,
    updateStatus,
    checkingUpdate,
    updatingPanel,
    formatBytes,
    getUsageClass,
    loadSystemInfo,
    loadSystemStats,
    loadVersion,
    handleCheckUpdate,
    handleUpdatePanel,
    handleRestartPanel,
    openGitHub
  }
}
