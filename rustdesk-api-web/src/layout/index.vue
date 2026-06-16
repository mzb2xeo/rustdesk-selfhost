<template>
  <el-config-provider :locale="appStore.setting.locale.value">
    <el-container :style="{'--sideBarWidth': sideBarWidth}">
      <el-aside :width="leftWidth" class="app-left">
        <g-aside></g-aside>
      </el-aside>
      <el-container class="app-container">
        <el-header class="app-header">
          <g-header></g-header>
        </el-header>
        <div class="header-tags">
          <tags></tags>
        </div>

        <el-main class="app-main">
          <router-view v-slot="{ Component }">
            <transition mode="out-in" name="el-fade-in-linear">
              <keep-alive :include="cachedTags">
                <component :is="Component"/>
              </keep-alive>
            </transition>
          </router-view>
        </el-main>
      </el-container>
    </el-container>
  </el-config-provider>
</template>

<script setup>
  import { useAppStore } from '@/store/app'
  import { useTagsStore } from '@/store/tags'
  import { ref, computed } from 'vue'
  import Tags from '@/layout/components/tags/index.vue'
  import GAside from '@/layout/components/aside.vue'
  import GHeader from '@/layout/components/header.vue'

  const appStore = useAppStore()
  const tagStore = useTagsStore()
  const sideBarWidth = computed(() => appStore.setting.locale.sideBarWidth)
  const leftWidth = computed(() => appStore.setting.sideIsCollapse ? '64px' : 'var(--sideBarWidth)')

  const cachedTags = ref([])

  cachedTags.value = tagStore.cached
</script>

<style lang="scss" scoped>
.app-header {
  background-color: var(--header-bg-color);
  color: var(--header-text-color);
  border-bottom: 1px solid var(--header-border-color);
  display: flex;
  height: 58px;
  position: sticky;
  top: 0;
  z-index: 20;
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
}

.header-tags {
  min-height: 42px;
  border-bottom: 1px solid var(--header-border-color);
  background-color: var(--header-bg-color);
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 18px;
  position: sticky;
  top: 58px;
  z-index: 19;
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  overflow-x: auto;
}

.app-left {
  transition: width 0.25s ease;
  border-right: 1px solid var(--side-border-color);
  background: var(--side-bg-color);
  position: sticky;
  top: 0;
  height: 100dvh;
  z-index: 30;
}

.app-container {
  min-height: 100dvh;
}

.app-main {
  min-height: calc(100dvh - 100px);
  background: transparent;
}
</style>


