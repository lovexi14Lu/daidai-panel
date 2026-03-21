<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import { systemApi } from '@/api/system'
import {
  Bell,
  Box,
  Connection,
  Document,
  Download,
  Expand,
  Fold,
  Key,
  Moon,
  Odometer,
  Operation,
  SetUp,
  Setting,
  Sunny,
  Tickets,
  Timer,
  User,
  UserFilled,
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const themeStore = useThemeStore()
const isCollapsed = ref(false)
const isMobile = ref(false)
const drawerVisible = ref(false)
const panelTitle = ref('呆呆面板')
const panelIcon = ref('')
const panelVersion = ref('')

const roleLevel: Record<string, number> = {
  viewer: 1,
  operator: 2,
  admin: 3,
}

function hasRole(minRole: string) {
  const currentRole = authStore.user?.role
  if (!currentRole) return false
  return (roleLevel[currentRole] || 0) >= (roleLevel[minRole] || 0)
}

const currentSection = computed(() => {
  const matched = [...route.matched].reverse().find(item => item.meta.section)
  return matched?.meta.section === 'admin' ? 'admin' : 'workspace'
})

const canAccessAdmin = computed(() => hasRole('admin'))

const workspaceItems = [
  { index: '/dashboard', title: '仪表板', icon: Odometer, minRole: 'viewer' },
  { index: '/tasks', title: '定时任务', icon: Timer, minRole: 'viewer' },
  { index: '/subscriptions', title: '订阅管理', icon: Download, minRole: 'operator' },
  { index: '/envs', title: '环境变量', icon: Setting, minRole: 'operator' },
  { index: '/logs', title: '执行日志', icon: Tickets, minRole: 'viewer' },
  { index: '/scripts', title: '脚本管理', icon: Document, minRole: 'operator' },
  { index: '/deps', title: '依赖管理', icon: Box, minRole: 'admin' },
  { index: '/api-docs', title: '接口文档', icon: Connection, minRole: 'viewer' },
  { index: '/profile', title: '个人设置', icon: User, minRole: 'viewer' },
]

const adminItems = [
  { index: '/admin/settings', title: '系统设置', icon: SetUp, minRole: 'admin' },
  { index: '/admin/notifications', title: '通知渠道', icon: Bell, minRole: 'admin' },
  { index: '/admin/users', title: '用户管理', icon: UserFilled, minRole: 'admin' },
  { index: '/admin/open-api', title: 'Open API', icon: Key, minRole: 'admin' },
]

function checkMobile() {
  isMobile.value = window.innerWidth <= 768
  if (isMobile.value) isCollapsed.value = true
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  loadPanelSettings()
  loadVersion()
  if (authStore.isLoggedIn && !authStore.user) {
    authStore.fetchUser()
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})

const menuItems = computed(() => {
  const allItems = currentSection.value === 'admin' ? adminItems : workspaceItems
  return allItems.filter(item => hasRole(item.minRole))
})

const activeMenu = computed(() => route.path)
const sectionLabel = computed(() => currentSection.value === 'admin' ? '管理后台' : '工作台')
const sectionActionLabel = computed(() => currentSection.value === 'admin' ? '返回工作台' : '进入后台')
const sidebarToggleIcon = computed(() => (isMobile.value ? Operation : (isCollapsed.value ? Expand : Fold)))
const themeIcon = computed(() => (themeStore.isDark ? Sunny : Moon))

function handleMenuSelect(index: string) {
  router.push(index)
  if (isMobile.value) drawerVisible.value = false
}

function toggleSidebar() {
  if (isMobile.value) {
    drawerVisible.value = !drawerVisible.value
  } else {
    isCollapsed.value = !isCollapsed.value
  }
}

async function handleLogout() {
  await authStore.logout()
}

function toggleSection() {
  if (!canAccessAdmin.value) return
  router.push(currentSection.value === 'admin' ? '/dashboard' : '/admin/settings')
  if (isMobile.value) drawerVisible.value = false
}

async function loadPanelSettings() {
  try {
    const res = await systemApi.panelSettings() as any
    if (res.data?.panel_title) panelTitle.value = res.data.panel_title
    if (res.data?.panel_icon) panelIcon.value = res.data.panel_icon
  } catch {}
}

async function loadVersion() {
  try {
    const res = await systemApi.version() as any
    if (res.data?.version) panelVersion.value = res.data.version
  } catch {}
}
</script>

<template>
  <el-container class="layout-container">
    <el-aside v-if="!isMobile" :width="isCollapsed ? '64px' : '220px'" class="layout-aside">
      <div class="logo-area">
        <img :src="panelIcon || '/favicon.svg'" alt="logo" class="logo-img" />
        <span v-show="!isCollapsed" class="logo-text">{{ panelTitle }}</span>
        <span v-show="!isCollapsed && panelVersion" class="version-badge">v{{ panelVersion }}</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapsed"
        :collapse-transition="false"
        background-color="transparent"
        @select="handleMenuSelect"
      >
        <el-menu-item v-for="item in menuItems" :key="item.index" :index="item.index">
          <el-icon><component :is="item.icon" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-drawer
      v-if="isMobile"
      v-model="drawerVisible"
      direction="ltr"
      :size="240"
      :with-header="false"
      :show-close="false"
    >
      <div class="logo-area mobile-logo">
        <img :src="panelIcon || '/favicon.svg'" alt="logo" class="logo-img" />
        <span class="logo-text">{{ panelTitle }}</span>
        <span v-if="panelVersion" class="version-badge">v{{ panelVersion }}</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        background-color="transparent"
        @select="handleMenuSelect"
      >
        <el-menu-item v-for="item in menuItems" :key="item.index" :index="item.index">
          <el-icon><component :is="item.icon" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>
    </el-drawer>

    <el-container>
      <el-header class="layout-header">
        <div class="header-left">
          <el-button :icon="sidebarToggleIcon" text @click="toggleSidebar" />
          <span v-if="isMobile" class="mobile-title">{{ panelTitle }}</span>
          <el-tag size="small" effect="plain" class="section-tag">{{ sectionLabel }}</el-tag>
        </div>
        <div class="header-right">
          <el-button v-if="canAccessAdmin" text class="section-switch-btn" @click="toggleSection">
            {{ sectionActionLabel }}
          </el-button>
          <el-button :icon="themeIcon" text circle class="theme-btn" @click="themeStore.toggleTheme" />
          <el-dropdown trigger="click">
            <span class="user-dropdown">
              <el-icon><User /></el-icon>
              <span v-if="!isMobile">{{ authStore.user?.username || 'User' }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="router.push('/profile')">个人设置</el-dropdown-item>
                <el-dropdown-item v-if="canAccessAdmin" @click="router.push('/admin/settings')">系统设置</el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="layout-main">
        <router-view v-slot="{ Component, route: viewRoute }">
          <keep-alive :max="3">
            <component :is="Component" :key="viewRoute.path" />
          </keep-alive>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped lang="scss">
.layout-container {
  height: 100vh;
}

.layout-aside {
  border-right: 1px solid var(--el-border-color-light);
  transition: width 0.25s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color);
  will-change: width;
}

.logo-area {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-bottom: 1px solid var(--el-border-color-light);
  flex-shrink: 0;
  padding: 0 16px;

  .logo-img {
    width: 32px;
    height: 32px;
    transition: transform 0.3s cubic-bezier(0.68, -0.55, 0.265, 1.55);
    will-change: transform;

    &:hover {
      transform: rotate(10deg) scale(1.1);
    }
  }

  .logo-text {
    font-size: 18px;
    font-weight: 700;
    white-space: nowrap;
    background: linear-gradient(135deg, #409EFF, #7B5CFA);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .version-badge {
    font-size: 10px;
    font-weight: 600;
    padding: 1px 6px;
    border-radius: 8px;
    background: linear-gradient(135deg, #409EFF, #7B5CFA);
    color: #fff;
    white-space: nowrap;
    line-height: 1.4;
    letter-spacing: 0.3px;
    flex-shrink: 0;
    align-self: center;
  }
}

.mobile-logo {
  border-bottom: 1px solid var(--el-border-color-light);
  justify-content: flex-start;
  padding: 0 16px;
}

.mobile-title {
  font-size: 16px;
  font-weight: 600;
  margin-left: 4px;
  background: linear-gradient(135deg, #409EFF, #7B5CFA);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.layout-header {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--el-border-color-light);
  padding: 0 20px;
  background: var(--el-bg-color);
  backdrop-filter: blur(8px);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-tag {
  border-radius: 999px;
  padding: 0 10px;
}

.section-switch-btn {
  font-weight: 600;
  border-radius: 999px;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 8px;
  transition: all 0.2s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  font-weight: 500;

  &:hover {
    background: var(--el-fill-color-light);
    color: var(--el-color-primary);
    transform: translateY(-1px);
  }
}

.layout-main {
  background: var(--el-bg-color-page);
  overflow-y: auto;
  padding: 20px;
}

@media screen and (max-width: 768px) {
  .layout-header {
    padding: 0 12px;
    height: 50px;
  }

  .layout-main {
    padding: 12px;
  }
}

.theme-btn {
  transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1) !important;

  &:hover {
    transform: rotate(20deg) scale(1.15) !important;
  }

  &:active {
    transform: rotate(0deg) scale(0.95) !important;
  }
}

.page-fade-enter-active {
  animation: pageEnter 0.22s ease-out;
}

.page-fade-leave-active {
  animation: pageFadeOut 0.12s ease-in;
}

@keyframes pageEnter {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes pageFadeOut {
  from { opacity: 1; }
  to { opacity: 0; }
}
</style>
