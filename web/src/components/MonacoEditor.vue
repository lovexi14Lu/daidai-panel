<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import type * as MonacoType from 'monaco-editor'
import { loadMonacoEditor } from '@/utils/monaco'

const props = defineProps<{
  modelValue: string
  language?: string
  readonly?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const editorRef = ref<HTMLElement>()
const isLoading = ref(true)
const loadError = ref('')
let editor: MonacoType.editor.IStandaloneCodeEditor | null = null
let monacoInstance: typeof MonacoType | null = null

onMounted(async () => {
  if (!editorRef.value) return

  try {
    loadError.value = ''
    const { monaco, source } = await loadMonacoEditor()
    monacoInstance = monaco as typeof MonacoType
    if (!editorRef.value) return

    editor = monacoInstance.editor.create(editorRef.value, {
      value: props.modelValue,
      language: props.language || 'javascript',
      theme: 'vs-dark',
      automaticLayout: true,
      fontSize: 14,
      minimap: { enabled: true },
      scrollBeyondLastLine: false,
      readOnly: props.readonly || false,
      tabSize: 2,
      wordWrap: 'on',
    })

    if (source === 'cdn') {
      console.warn('Monaco 编辑器当前已回退到 CDN 资源。')
    }

    editor!.onDidChangeModelContent(() => {
      if (editor) {
        emit('update:modelValue', editor.getValue())
      }
    })
  } catch (error) {
    console.error('Monaco 编辑器初始化失败', error)
    loadError.value = '编辑器加载失败，请检查网络或稍后重试。'
  } finally {
    isLoading.value = false
  }
})

watch(() => props.modelValue, (newValue) => {
  if (editor && newValue !== editor.getValue()) {
    editor.setValue(newValue)
  }
})

watch(() => props.language, (newLang) => {
  if (editor && monacoInstance) {
    const model = editor.getModel()
    if (model) {
      monacoInstance.editor.setModelLanguage(model, newLang || 'javascript')
    }
  }
})

watch(() => props.readonly, (newReadonly) => {
  if (editor) {
    editor.updateOptions({ readOnly: newReadonly })
  }
})

onBeforeUnmount(() => {
  editor?.dispose()
  editor = null
})

defineExpose({
  format: () => {
    if (editor) {
      editor.getAction('editor.action.formatDocument')?.run()
    }
  },
  getValue: () => editor?.getValue() || '',
  setValue: (value: string) => editor?.setValue(value),
})
</script>

<template>
  <div class="monaco-editor-wrapper">
    <div v-if="isLoading" class="monaco-loading">
      <div class="loading-spinner"></div>
      <span>编辑器加载中...</span>
    </div>
    <div v-else-if="loadError" class="monaco-loading monaco-error">
      <span>{{ loadError }}</span>
    </div>
    <div ref="editorRef" class="monaco-editor-container" v-show="!isLoading && !loadError"></div>
  </div>
</template>

<style scoped>
.monaco-editor-wrapper {
  width: 100%;
  height: 100%;
  min-height: 400px;
  position: relative;
}

.monaco-editor-container {
  width: 100%;
  height: 100%;
}

.monaco-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 400px;
  gap: 12px;
  color: var(--el-text-color-secondary);
  font-size: 14px;
  background: #1e1e1e;
  border-radius: 4px;
}

.monaco-error {
  color: #f56c6c;
  text-align: center;
}

.loading-spinner {
  width: 28px;
  height: 28px;
  border: 3px solid rgba(255, 255, 255, 0.15);
  border-top-color: #409eff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
